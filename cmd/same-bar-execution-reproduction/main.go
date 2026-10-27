package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"

	"research-platform/handoff"
	"research-platform/seal"
	"research-platform/spec"
	"same-bar-execution-reproduction/study"
)

const producer = "same-bar-execution-reproduction"

var result = seal.OutputDeclaration{Path: "result.json", Role: "result", MediaType: "application/json", Producer: producer}
var failure = seal.OutputDeclaration{Path: "failure.json", Role: "technical_failure", MediaType: "application/json", Producer: producer}
var semantic = seal.OutputDeclaration{Path: "semantic.json", Role: "semantic_verification", MediaType: "application/json", Producer: producer}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 2 || len(args) > 3 {
		return fmt.Errorf("usage: same-bar-execution-reproduction run <lineage> | verify <run-id> | handoff <lineage> <run-id>")
	}
	w, err := seal.OpenWorkspace(".")
	if err != nil {
		return err
	}
	defer w.Close()
	switch args[0] {
	case "run":
		if len(args) == 2 {
			return execute(w, args[1])
		}
	case "verify":
		if len(args) == 2 {
			v, s, err := verify(w, args[1])
			if err != nil {
				return err
			}
			fmt.Println(v.RunID, "integrity PASS; semantic", s.Status)
			if s.Status == "FAIL" {
				return fmt.Errorf("semantic verification failed")
			}
			return nil
		}
	case "handoff":
		if len(args) == 3 {
			v, s, err := verify(w, args[2])
			if err != nil {
				return err
			}
			if s.Status == "FAIL" {
				return fmt.Errorf("semantic verification failed")
			}
			_, b, err := handoff.Build(context.Background(), w, args[1], v.RunID, s)
			if err != nil {
				return err
			}
			if err = os.MkdirAll("handoffs", 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(path.Join("handoffs", v.RunID+".json"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
			if err != nil {
				return err
			}
			if _, err = f.Write(b); err == nil {
				err = f.Sync()
			}
			if closeErr := f.Close(); err == nil {
				err = closeErr
			}
			if err != nil {
				return err
			}
			fmt.Println("handoffs/" + v.RunID + ".json")
			return nil
		}
	}
	return fmt.Errorf("usage: same-bar-execution-reproduction run <lineage> | verify <run-id> | handoff <lineage> <run-id>")
}

func execute(w *seal.Workspace, lineageID string) error {
	if _, err := w.LoadLineage(lineageID); err != nil {
		return err
	}
	rules := []seal.TerminalOutputRule{
		{Outcome: seal.Complete, RequiredOutputs: []seal.OutputDeclaration{result, semantic}},
		{Outcome: seal.Incomplete, RequiredOutputs: []seal.OutputDeclaration{result, semantic}},
		{Outcome: seal.TechnicalFailure, RequiredOutputs: []seal.OutputDeclaration{failure}, ForbiddenRoles: []string{"result", "semantic_verification"}},
	}
	r, err := w.BeginRun(context.Background(), lineageID, seal.RunOptions{OutputRules: rules, PermittedOutcomes: []seal.TerminalOutcome{seal.Complete, seal.Incomplete, seal.TechnicalFailure}})
	if err != nil {
		return err
	}
	fmt.Println("run:", r.ID())
	in := study.NewInputs(w, r.SourceData())
	e, err := study.Evaluate(in)
	if err != nil {
		return err
	}
	if len(e.Payload) == 0 {
		return fmt.Errorf("empty study payload")
	}
	if e.Outcome == seal.TechnicalFailure {
		if _, err := publish(r, failure, e.Payload); err != nil {
			return err
		}
	} else if e.Outcome == seal.Complete || e.Outcome == seal.Incomplete {
		if _, err := publish(r, result, e.Payload); err != nil {
			return err
		}
		v, err := study.Verify(in, e.Payload)
		if err != nil {
			return err
		}
		if v.Status != "PASS" && v.Status != "FAIL" {
			return fmt.Errorf("invalid semantic status")
		}
		b, err := spec.CanonicalJSON(v)
		if err != nil {
			return err
		}
		ref, err := publish(r, semantic, b)
		if err != nil {
			return err
		}
		_, err = r.Seal(context.Background(), seal.SealRequest{Outcome: e.Outcome, SemanticVerification: &ref})
		return err
	} else {
		return fmt.Errorf("invalid terminal outcome")
	}
	_, err = r.Seal(context.Background(), seal.SealRequest{Outcome: e.Outcome})
	return err
}

func publish(r *seal.Run, d seal.OutputDeclaration, b []byte) (seal.ArtifactRef, error) {
	return r.PublishImmutable(context.Background(), d.Path, bytes.NewReader(b), seal.ArtifactDeclaration{Role: d.Role, MediaType: d.MediaType, Producer: d.Producer})
}

func verify(w *seal.Workspace, id string) (seal.VerifiedRun, handoff.SemanticStatus, error) {
	v, err := w.VerifyIntegrity(context.Background(), id)
	if err != nil {
		return v, handoff.SemanticStatus{}, err
	}
	implementationNotice(v.StudyImplementationSHA256, v.StudyGitCommit)
	if v.Outcome == seal.TechnicalFailure {
		return v, handoff.SemanticStatus{Status: "NOT_APPLICABLE"}, nil
	}
	var resultRef seal.ArtifactRef
	for _, a := range v.Artifacts {
		if a.Role == "result" {
			resultRef = a
		}
	}
	if resultRef.Path != "result.json" || v.SemanticVerification == nil {
		return v, handoff.SemanticStatus{}, fmt.Errorf("missing sealed result or semantic record")
	}
	b, err := w.ReadArtifact(id, resultRef, 8<<20)
	if err != nil {
		return v, handoff.SemanticStatus{}, err
	}
	got, err := study.Verify(study.NewInputs(w, v.SourceData), b)
	if err != nil {
		return v, handoff.SemanticStatus{}, err
	}
	if got.Status != "PASS" && got.Status != "FAIL" {
		return v, handoff.SemanticStatus{}, fmt.Errorf("invalid semantic status")
	}
	want, err := spec.CanonicalJSON(got)
	if err != nil {
		return v, handoff.SemanticStatus{}, err
	}
	stored, err := w.ReadArtifact(id, *v.SemanticVerification, 8<<20)
	if err != nil {
		return v, handoff.SemanticStatus{}, err
	}
	if !bytes.Equal(want, stored) {
		return v, handoff.SemanticStatus{}, fmt.Errorf("semantic verification differs from sealed record")
	}
	return v, handoff.SemanticStatus{Status: got.Status, Artifact: *v.SemanticVerification}, nil
}

func implementationNotice(sealed, commit string) {
	current, err := seal.StudyImplementationSHA256(".")
	if err != nil {
		fmt.Println("study implementation: unavailable (" + err.Error() + ")")
	} else if current == sealed {
		fmt.Println("study implementation: matches sealed")
	} else {
		fmt.Printf("study implementation: differs from sealed (sealed %s, current %s; replay at commit %s)\n", sealed, current, commit)
	}
}
