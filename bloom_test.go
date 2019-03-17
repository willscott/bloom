// bloom_test.go - Bloom filter tests.
// Written in 2017 by Yawning Angel
//
// To the extent possible under law, the author(s) have dedicated all copyright
// and related and neighboring rights to this software to the public domain
// worldwide. This software is distributed without any warranty.
//
// You should have received a copy of the CC0 Public Domain Dedication along
// with this software. If not, see <http://creativecommons.org/publicdomain/zero/1.0/>.

package bloom

import (
	"bytes"
	"compress/zlib"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	const (
		entryLength       = 32
		filterSize        = 15 // 2^15 bits = 4 KiB

		expectedEntries = 1024
	)

	assert := assert.New(t)
	require := require.New(t)

	// 4 KiB filter, 1/2^5 (32 bit) load
	f, err := New(rand.Reader, filterSize, .03125)
	require.NoError(err, "New()")
	assert.Equal(0, f.Entries(), "Entries(), empty filter")

	// Assert that the bloom filter math is correct.
	assert.Equal(expectedEntries, f.MaxEntries(), "Max entries")

	// Generate enough entries to fully saturate the filter.
	max := f.MaxEntries()
	entries := make(map[[entryLength]byte]bool)
	for count := 0; count < max; {
		var ent [entryLength]byte
		rand.Read(ent[:])

		// This needs to ignore false positives.
		if !f.TestAndSet(ent[:]) {
			entries[ent] = true
			count++
		}
	}
	assert.Equal(max, f.Entries(), "After populating")

	// Ensure that all the entries are present in the filter.
	idx := 0
	for ent := range entries {
		assert.True(f.Test(ent[:]), "Test(ent #: %v)", idx)
		assert.True(f.TestAndSet(ent[:]), "TestAndSet(ent #: %v)", idx)
		idx++
	}

	// Test the false positive rate, by generating another set of entries
	// NOT in the filter, and counting the false positives.
	//
	// This may have suprious failures once in a blue moon because the
	// algorithm is probabalistic, but that's *exceedingly* unlikely with
	// the chosen delta.
	randomEntries := make(map[[entryLength]byte]bool)
	for count := 0; count < max; {
		var ent [entryLength]byte
		rand.Read(ent[:])
		if !entries[ent] && !randomEntries[ent] {
			randomEntries[ent] = true
			count++
		}
	}
	falsePositives := 0
	for ent := range randomEntries {
		if f.Test(ent[:]) {
			falsePositives++
		}
	}
	observedP := float64(falsePositives) / float64(max)
	t.Logf("Observed False Positive Rate: %v", observedP)
	//assert.Lessf(observedP, 0.02, "False positive rate")

	assert.Equal(max, f.Entries(), "After tests") // Should still be = max.
}

func TestFilterCompression(t *testing.T) {
	const (
		entryLength       = 32
		entries        = 20 // 2^20 bits = 1m bits
		clientcnt = 1000
	)

	assert := assert.New(t)
	require := require.New(t)

	for _, epochs := range []int{1, 5, 10, 20} {

	for i, load := range []int{1,2,4,8,16,32} {
		l := float64(1) / float64(load)
		f, err := New(rand.Reader, entries + i, l)
		require.NoError(err, "New()")
		assert.Equal(0, f.Entries(), "Entries(), empty filter")	
		fmt.Printf("%d layer; bits/el %d: ", epochs*clientcnt, load)

		entries := make(map[[entryLength]byte]bool)
		for count := 0; count < epochs * clientcnt; {
			var ent [entryLength]byte
			rand.Read(ent[:])
	
			// This needs to ignore false positives.
			if !f.TestAndSet(ent[:]) {
				entries[ent] = true
				count++
			}
		}
		layer := f.Delta()
		var b bytes.Buffer
		w := zlib.NewWriter(&b)
		w.Write(layer)
		w.Close()

		// we now have a compressed size.
		fmt.Printf("%d bytes. ", len(b.Bytes()))
		// Now try a bunch of requests to understand the false positive rate.
		randomEntries := make(map[[entryLength]byte]bool)
		for count := 0; count < clientcnt * 100; {
			var ent [entryLength]byte
			rand.Read(ent[:])
			if !entries[ent] && !randomEntries[ent] {
				randomEntries[ent] = true
				count++
			}
		}
		falsePositives := 0
		for ent := range randomEntries {
			if f.Test(ent[:]) {
				falsePositives++
			}
		}
		observedP := float64(falsePositives) / float64(clientcnt * 100)
		fmt.Printf("fp: %f\n", observedP)
	}
	}
}
