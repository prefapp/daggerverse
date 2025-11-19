package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"strings"
	"fmt"
	"time"
)

func isOperatorUp(

    ctx context.Context,
    kindContainer *dagger.Container,


) (bool, error) {

	namespace := "default"
	labelSelector := "release=firestartr-init"
	
	phaseTemplate := `{{range .items}}{{.status.phase}}{{end}}`

    getPhaseCommand := []string{
        "kubectl", 
        "get", 
        "pod", 
        "-l", // Use label selector
        labelSelector, 
        "-n", 
        namespace,
        "-o", 
        fmt.Sprintf("go-template=%s", phaseTemplate),
    }

	output, err := kindContainer.
        WithExec(getPhaseCommand).
        Stdout(ctx)

	if err != nil {
        errorOutput, _ := kindContainer.Stderr(ctx)
        return false, fmt.Errorf("kubectl failed to get pod status: %s", strings.TrimSpace(errorOutput))
    }	

	podPhase := strings.TrimSpace(output)

	if podPhase == "" {
        fmt.Printf("No Pod found matching selector '%s' in namespace '%s'.\n", labelSelector, namespace)
        return false, nil // Return false if not found
    }

    fmt.Printf("Found Pod phase: %s\n", podPhase)

    if podPhase == "Running" {
        return true, nil
    } 
    
    return false, nil
}

func (m *FirestartrBootstrap) ProcessArtifactsByKind(
    ctx context.Context,
    kindContainer *dagger.Container,
) error {

	namespace := "default"

	running, err := isOperatorUp(ctx, kindContainer)

	if err != nil {
		panic(err)
	}

	if running != true {
		return fmt.Errorf("The operator is not up")
	}

	var targetKinds = []string{
		"githubrepositoryfeature",
		"githubrepositorysecretssections",
		"githubrepository",
		"githuborgwebhook",
		"githubgroup",
	}

	for _, kind := range targetKinds {

		var artifacts []string

		if (kind == "githubrepositoryfeature" || 
		kind == "githubrepositorysecretssections") {
			artifacts, err = m.GetArtifactListByKind(
				ctx,
				kindContainer,
				kind,
				namespace,
			)

		}else{
			// only the artifacts with the annotation (to avoid imported items)
			artifacts, err = m.GetAnnotatedArtifactList(
				ctx,
				kindContainer,
				kind,
				namespace,
				"firestartr.dev/bootstrapped",
				"true",
			)
		}

        fmt.Println("-%s",strings.Join(artifacts, "\n-"))
	//	summary, err := m.DeleteArtifactsList(
	//		ctx, 
	//		kindContainer,
	//		kind,
	//		artifacts,
	//		namespace,
	//	)

	//	if err != nil {
	//		panic(err)
	//	}

	//	fmt.Println(summary)
	}

	return nil
}

func (m *FirestartrBootstrap) DeleteArtifactsList(

    ctx context.Context,
	kindContainer *dagger.Container,
	kind string,
	artifacts []string,
	namespace string,
) (string, error) {

	var summary strings.Builder

	timeoutSeconds := 180

	for _, artifactName := range artifacts {

        resourceRef := fmt.Sprintf("%s/%s", kind, artifactName)
        
        // Step A: Execute the Delete command
        deleteCmd := []string{
            "kubectl", 
            "delete", 
            resourceRef, 
            "-n", 
            namespace, 
            "--ignore-not-found=true", // Ensure pipeline doesn't break if already gone
            "--wait=false",            // Don't block here, we use a separate wait command below
        }
        
        _, err := kindContainer.WithExec(deleteCmd).Stdout(ctx)
        if err != nil {
            // Log the error but continue to the next artifact if possible
            fmt.Fprintf(&summary, "  ERROR during kubectl delete: %v. Continuing...\n", err)
            continue
        }
        
        // Step B: Execute the Wait command
        // The core requirement: wait for the resource to be effectively deleted by the operator.
        // This command exits successfully only when the resource is gone.
        waitCmd := []string{
            "kubectl", 
            "wait", 
            "--for=delete", // Wait until the resource is gone
            resourceRef, 
            "-n", 
            namespace,
            fmt.Sprintf("--timeout=%ds", timeoutSeconds), // Timeout for this specific wait
        }
        
        start := time.Now()
        _, err = kindContainer.WithExec(waitCmd).Stdout(ctx)
        duration := time.Since(start)

        if err != nil {
            // Capture stderr for better debugging
            errorOutput, _ := kindContainer.Stderr(ctx)
            
            // If the error is *not* a timeout (which is expected if the operator fails), 
            // we should log a specific failure.
            if !strings.Contains(errorOutput, "timed out waiting for the condition") {
                fmt.Fprintf(&summary, "  ERROR during kubectl wait: Resource may be stuck. Failed with: %s\n", strings.TrimSpace(errorOutput))
            } else {
                 fmt.Fprintf(&summary, "  WAIT TIMEOUT: Resource not deleted within %d seconds. It may be stuck or the operator failed cleanup.\n", timeoutSeconds)
            }
            // Decide if we should continue to the next or exit. We continue to clean up others.
            continue
        }

        fmt.Fprintf(&summary, "  SUCCESS: %s confirmed deleted in %s.\n", resourceRef, duration.Round(time.Second))
    }

    return summary.String(), nil
}

func (m *FirestartrBootstrap) GetArtifactListByKind(
    ctx context.Context,
	kindContainer *dagger.Container,
    resourceKind string,
    namespace string,
) ([]string, error) {
	
    // JSONPath: .items[*].metadata.name extracts the name field from every item in the list.
    jsonPathFilter := "jsonpath='{.items[*].metadata.name}'"
    
    getCommand := []string{
        "kubectl", 
        "get", 
        resourceKind, 
        "-n", 
        namespace,
        "-o", 
        jsonPathFilter, 
    }

    output, err := kindContainer.
        WithExec(getCommand).
        Stdout(ctx)
        
    if err != nil {
        errorOutput, _ := kindContainer.Stderr(ctx)
        return nil, fmt.Errorf("kubectl command failed: %s", strings.TrimSpace(errorOutput))
    }

    rawOutput := strings.TrimSpace(output)
    
    if rawOutput == "" {
        return []string{}, nil
    }

    artifactList := strings.Fields(rawOutput) 
    
    return artifactList, nil
}

func (m *FirestartrBootstrap) GetAnnotatedArtifactList(
    ctx context.Context,
    kindContainer *dagger.Container,
    resourceKind string,
    namespace string,
    annotationKey string,
    annotationValue string,
) ([]string, error) {

	templateFilter := fmt.Sprintf(
        `{{range .items}}` +
        `{{if .metadata.annotations}}` +
        `{{if eq (index .metadata.annotations "%s") "%s"}}` +
        `{{.metadata.name}}{{"\n"}}` +
        `{{end}}` +
        `{{end}}` +
        `{{end}}`,
        annotationKey, annotationValue,
    )
    
    getAnnotatedCommand := []string{
        "kubectl", 
        "get", 
        resourceKind, 
        "-n", 
        namespace,
        "-o", 
        fmt.Sprintf("go-template=%s", templateFilter),
    }

	output, err := kindContainer.
        WithExec(getAnnotatedCommand).
        Stdout(ctx)

	if err != nil {
        errorOutput, _ := kindContainer.Stderr(ctx)
        return nil, fmt.Errorf("kubectl command failed. Error: %s", strings.TrimSpace(errorOutput))
    }

	rawOutput := strings.TrimSpace(output)
    
    if rawOutput == "" {
        // No matching artifacts found.
        return []string{}, nil
    }

    // Split the newline-separated output into a slice of strings
    artifactList := strings.Split(rawOutput, "\n")
    
    return artifactList, nil
}
