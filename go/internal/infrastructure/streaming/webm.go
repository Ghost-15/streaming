package streaming

import "bytes"

var webMClusterID = []byte{0x1f, 0x43, 0xb6, 0x75}

// webMInitializationPrefix returns the MSE WebM initialization segment: all
// bytes before the first Cluster element. Unlike a MediaRecorder timeslice
// blob, this prefix contains metadata only and therefore cannot replay audio.
func webMInitializationPrefix(data []byte) ([]byte, bool) {
	clusterAt := bytes.Index(data, webMClusterID)
	if clusterAt < 0 {
		return nil, false
	}
	return append([]byte(nil), data[:clusterAt]...), true
}

// WebMClusterAligner discards a late listener's partial current Cluster and
// resumes at the next complete Cluster ID. It retains three trailing bytes so
// the four-byte ID is still detected when split across two publisher chunks.
type WebMClusterAligner struct {
	tail    []byte
	aligned bool
}

// Align returns nil until the next WebM Cluster boundary, then returns a
// contiguous byte stream beginning with that boundary. Once aligned, data is
// forwarded unchanged and in order.
func (a *WebMClusterAligner) Align(data []byte) []byte {
	if a.aligned {
		return data
	}
	combined := make([]byte, 0, len(a.tail)+len(data))
	combined = append(combined, a.tail...)
	combined = append(combined, data...)
	if clusterAt := bytes.Index(combined, webMClusterID); clusterAt >= 0 {
		a.aligned = true
		a.tail = nil
		return combined[clusterAt:]
	}

	keep := len(webMClusterID) - 1
	if len(combined) < keep {
		keep = len(combined)
	}
	a.tail = append(a.tail[:0], combined[len(combined)-keep:]...)
	return nil
}
