package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestBuildNewImagesYAML(t *testing.T) {
	cases := []struct {
		name   string
		images []ImageData
		want   string
	}{
		{"empty", nil, "{}"},
		{"services", []ImageData{
			{ServiceNameList: []string{"micro-a", "micro-c"}, Image: "image-A"},
			{ServiceNameList: []string{"micro-b"}, Image: "image-B"},
		}, "micro-a: {image: image-A}\nmicro-b: {image: image-B}\nmicro-c: {image: image-A}"},
		{"shared parent keys", []ImageData{
			{ImageKeys: []string{"/workers/main/image", "/workers/main/sidecar/image"}, Image: "image-A"},
			{ImageKeys: []string{"/workers/other/image"}, Image: "image-B"},
		}, "workers: {main: {image: image-A, sidecar: {image: image-A}}, other: {image: image-B}}"},
		{"service then key", []ImageData{
			{ServiceNameList: []string{"micro-a"}, Image: "image-A"},
			{ImageKeys: []string{"/micro-a/sidecar/image"}, Image: "image-B"},
		}, "micro-a: {image: image-A, sidecar: {image: image-B}}"},
		{"key then service", []ImageData{
			{ImageKeys: []string{"/micro-a/sidecar/image"}, Image: "image-B"},
			{ServiceNameList: []string{"micro-a"}, Image: "image-A"},
		}, "micro-a: {image: image-A, sidecar: {image: image-B}}"},
		{"escaped keys and image", []ImageData{
			{ImageKeys: []string{"/a~1b/~0image"}, Image: "registry/image:\"tag"},
		}, "'a/b': {'~image': 'registry/image:\"tag'}"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input, err := json.Marshal(ImageMatrix{Images: tc.images})
			if err != nil {
				t.Fatal(err)
			}
			output, err := buildNewImagesYAML(string(input))
			if err != nil {
				t.Fatal(err)
			}
			var got, want map[string]interface{}
			if err := yaml.Unmarshal([]byte(output), &got); err != nil {
				t.Fatal(err)
			}
			if err := yaml.Unmarshal([]byte(tc.want), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %s, want %s", output, tc.want)
			}
		})
	}
}

func TestBuildNewImagesYAMLRejectsConflicts(t *testing.T) {
	cases := []struct {
		name   string
		images []ImageData
		want   string
	}{
		{"missing targets", []ImageData{{Image: "A"}}, "cannot be empty"},
		{"mixed entry", []ImageData{{ServiceNameList: []string{"a"}, ImageKeys: []string{"/b"}, Image: "A"}}, "cannot be used together"},
		{"duplicate service", []ImageData{{ServiceNameList: []string{"a", "a"}, Image: "A"}}, "duplicate service"},
		{"duplicate key", []ImageData{{ImageKeys: []string{"/a/image"}, Image: "A"}, {ImageKeys: []string{"/a/image"}, Image: "B"}}, "conflicting image target"},
		{"service then conflicting key", []ImageData{{ServiceNameList: []string{"a"}, Image: "A"}, {ImageKeys: []string{"/a/image"}, Image: "B"}}, "conflicting image target"},
		{"key then conflicting service", []ImageData{{ImageKeys: []string{"/a/image"}, Image: "B"}, {ServiceNameList: []string{"a"}, Image: "A"}}, "conflicting image target"},
		{"parent then child", []ImageData{{ImageKeys: []string{"/a", "/a/image"}, Image: "A"}}, "conflicting image target"},
		{"child then parent", []ImageData{{ImageKeys: []string{"/a/image", "/a"}, Image: "A"}}, "conflicting image target"},
		{"invalid pointer", []ImageData{{ImageKeys: []string{"a/image"}, Image: "A"}}, "starting with /"},
		{"invalid escape", []ImageData{{ImageKeys: []string{"/a/~2image"}, Image: "A"}}, "invalid JSON pointer escape"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input, err := json.Marshal(ImageMatrix{Images: tc.images})
			if err != nil {
				t.Fatal(err)
			}
			_, err = buildNewImagesYAML(string(input))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected error containing %q, got %v", tc.want, err)
			}
		})
	}
	if _, err := buildNewImagesYAML("{"); err == nil {
		t.Fatal("expected malformed JSON error")
	}
}
