package streaming

import (
	"bytes"
	"testing"
)

func TestBroadcastWithInitCachesMetadataWithoutClusterAudio(t *testing.T) {
	hub := NewHub()
	metadata := []byte{0x1a, 0x45, 0xdf, 0xa3, 0x18, 0x53, 0x80, 0x67}
	firstRecorderBlob := append(append([]byte{}, metadata...), webMClusterID...)
	firstRecorderBlob = append(firstRecorderBlob, []byte("old-audio")...)

	hub.BroadcastWithInit("stream-1", firstRecorderBlob)

	if got := hub.InitSegment("stream-1"); !bytes.Equal(got, metadata) {
		t.Fatalf("InitSegment() = %x, want metadata only %x", got, metadata)
	}
}

func TestBroadcastWithInitFindsClusterIDSplitAcrossChunks(t *testing.T) {
	hub := NewHub()
	metadata := []byte("webm-metadata")
	hub.BroadcastWithInit("stream-1", append(append([]byte{}, metadata...), webMClusterID[:2]...))
	if got := hub.InitSegment("stream-1"); got != nil {
		t.Fatalf("InitSegment() before complete Cluster ID = %x, want nil", got)
	}

	hub.BroadcastWithInit("stream-1", append(append([]byte{}, webMClusterID[2:]...), []byte("media")...))
	if got := hub.InitSegment("stream-1"); !bytes.Equal(got, metadata) {
		t.Fatalf("InitSegment() = %x, want %x", got, metadata)
	}
}

func TestWebMClusterAlignerWaitsForBoundaryAcrossChunks(t *testing.T) {
	var aligner WebMClusterAligner
	if got := aligner.Align([]byte{0xaa, 0xbb, 0x1f, 0x43}); got != nil {
		t.Fatalf("Align() before full Cluster ID = %x, want nil", got)
	}

	media := append(append([]byte{}, webMClusterID...), 0x81, 0x00)
	got := aligner.Align([]byte{0xb6, 0x75, 0x81, 0x00})
	if !bytes.Equal(got, media) {
		t.Fatalf("Align() = %x, want %x", got, media)
	}

	next := []byte("cluster-continuation")
	if got = aligner.Align(next); !bytes.Equal(got, next) {
		t.Fatalf("Align() after boundary = %x, want %x", got, next)
	}
}
