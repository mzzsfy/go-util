//nolint:all
//go:build amd64

package swiss

// MatchMetadata performs a 16-way probe of |metadata| using SSE instructions
// nb: |metadata| must be an aligned pointer
func MatchMetadata(metadata *[16]int8, hash int8) uint16
