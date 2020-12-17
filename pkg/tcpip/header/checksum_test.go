// Copyright 2019 The gVisor Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package header provides the implementation of the encoding and decoding of
// network protocol headers.
package header_test

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

func TestChecksumVVWithOffset(t *testing.T) {
	testCases := []struct {
		name      string
		vv        buffer.VectorisedView
		off, size int
		initial   uint16
		want      uint16
	}{
		{
			name: "empty",
			vv: buffer.NewVectorisedView(0, []buffer.View{
				buffer.NewViewFromBytes([]byte{1, 9, 0, 5, 4}),
			}),
			off:  0,
			size: 0,
			want: 0,
		},
		{
			name: "OneView",
			vv: buffer.NewVectorisedView(0, []buffer.View{
				buffer.NewViewFromBytes([]byte{1, 9, 0, 5, 4}),
			}),
			off:  0,
			size: 5,
			want: 1294,
		},
		{
			name: "TwoViews",
			vv: buffer.NewVectorisedView(0, []buffer.View{
				buffer.NewViewFromBytes([]byte{1, 9, 0, 5, 4}),
				buffer.NewViewFromBytes([]byte{4, 3, 7, 1, 2, 123}),
			}),
			off:  0,
			size: 11,
			want: 33819,
		},
		{
			name: "TwoViewsWithOffset",
			vv: buffer.NewVectorisedView(0, []buffer.View{
				buffer.NewViewFromBytes([]byte{98, 1, 9, 0, 5, 4}),
				buffer.NewViewFromBytes([]byte{4, 3, 7, 1, 2, 123}),
			}),
			off:  1,
			size: 11,
			want: 33819,
		},
		{
			name: "ThreeViewsWithOffset",
			vv: buffer.NewVectorisedView(0, []buffer.View{
				buffer.NewViewFromBytes([]byte{98, 1, 9, 0, 5, 4}),
				buffer.NewViewFromBytes([]byte{98, 1, 9, 0, 5, 4}),
				buffer.NewViewFromBytes([]byte{4, 3, 7, 1, 2, 123}),
			}),
			off:  7,
			size: 11,
			want: 33819,
		},
		{
			name: "ThreeViewsWithInitial",
			vv: buffer.NewVectorisedView(0, []buffer.View{
				buffer.NewViewFromBytes([]byte{77, 11, 33, 0, 55, 44}),
				buffer.NewViewFromBytes([]byte{98, 1, 9, 0, 5, 4}),
				buffer.NewViewFromBytes([]byte{4, 3, 7, 1, 2, 123, 99}),
			}),
			initial: 77,
			off:     7,
			size:    11,
			want:    33896,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got, want := header.ChecksumVVWithOffset(tc.vv, tc.initial, tc.off, tc.size), tc.want; got != want {
				t.Errorf("header.ChecksumVVWithOffset(%v) = %v, want: %v", tc, got, tc.want)
			}
			v := tc.vv.ToView()
			v.TrimFront(tc.off)
			v.CapLength(tc.size)
			if got, want := header.Checksum(v, tc.initial), tc.want; got != want {
				t.Errorf("header.Checksum(%v) = %v, want: %v", tc, got, tc.want)
			}
		})
	}
}

func TestChecksum(t *testing.T) {
	var bufSizes = []int{0, 1, 2, 3, 4, 7, 8, 15, 16, 31, 32, 63, 64, 127, 128, 255, 256, 257, 1023, 1024}
	type testCase struct {
		buf      []byte
		initial  uint16
		csumOrig uint16
		csumNew  uint16
	}
	testCases := make([]testCase, 100000)
	// Ensure same buffer generation for test consistency.
	rnd := rand.New(rand.NewSource(42))
	for i := range testCases {
		testCases[i].buf = make([]byte, bufSizes[i%len(bufSizes)])
		testCases[i].initial = uint16(rnd.Intn(65536))
		rnd.Read(testCases[i].buf)
	}

	for i := range testCases {
		testCases[i].csumOrig = header.ChecksumOld(testCases[i].buf, testCases[i].initial)
		testCases[i].csumNew = header.Checksum(testCases[i].buf, testCases[i].initial)
		if got, want := testCases[i].csumNew, testCases[i].csumOrig; got != want {
			t.Fatalf("new checksum for (buf = %x, initial = %d) does not match old got: %d, want: %d", testCases[i].buf, testCases[i].initial, got, want)
		}
	}
}

func BenchmarkChecksum(b *testing.B) {
	var bufSizes = []int{64, 128, 256, 512, 1024, 1500, 2048, 4096, 8192, 16384, 32767, 32768, 65535, 65536}

	checkSumImpls := []struct {
		fn   func([]byte, uint16) uint16
		name string
	}{
		{header.ChecksumOld, fmt.Sprintf("checksum_old")},
		{header.Checksum, fmt.Sprintf("checksum")},
	}

	for _, csumImpl := range checkSumImpls {
		// Ensure same buffer generation for test consistency.
		rnd := rand.New(rand.NewSource(42))
		for _, bufSz := range bufSizes {
			b.Run(fmt.Sprintf("%s_%d", csumImpl.name, bufSz), func(b *testing.B) {
				tc := struct {
					buf     []byte
					initial uint16
					csum    uint16
				}{
					buf:     make([]byte, bufSz),
					initial: uint16(rnd.Intn(65536)),
				}
				rnd.Read(tc.buf)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					tc.csum = csumImpl.fn(tc.buf, tc.initial)
				}
			})
		}
	}
}

func testICMPChecksum(t *testing.T, headerChecksum func() uint16, icmpChecksum func() uint16, want uint16, pktStr string) {
	// icmpChecksum should not do any modifications of the header to
	// calculate its checksum. Let's call it from a few go-routines and the
	// race detector will trigger a warning if there are any concurrent
	// read/write accesses.

	const concurrency = 5
	start := make(chan int)
	ready := make(chan bool, concurrency)
	var wg sync.WaitGroup
	wg.Add(concurrency)
	defer wg.Wait()

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()

			ready <- true
			<-start

			if got := headerChecksum(); want != got {
				t.Errorf("new checksum for %s does not match old got: %x, want: %x", pktStr, got, want)
			}
			if got := icmpChecksum(); want != got {
				t.Errorf("new checksum for %s does not match old got: %x, want: %x", pktStr, got, want)
			}
		}()
	}
	for i := 0; i < concurrency; i++ {
		<-ready
	}
	close(start)
}

func TestICMPv4Checksum(t *testing.T) {
	rnd := rand.New(rand.NewSource(42))

	h := header.ICMPv4(make([]byte, header.ICMPv4MinimumSize))
	if _, err := rnd.Read(h); err != nil {
		t.Fatalf("rnd.Read failed: %v", err)
	}
	h.SetChecksum(0)

	buf := make([]byte, 13)
	if _, err := rnd.Read(buf); err != nil {
		t.Fatalf("rnd.Read failed: %v", err)
	}
	vv := buffer.NewVectorisedView(len(buf), []buffer.View{
		buffer.NewViewFromBytes(buf[:5]),
		buffer.NewViewFromBytes(buf[5:]),
	})

	want := header.Checksum(vv.ToView(), 0)
	want = ^header.Checksum(h, want)
	h.SetChecksum(want)

	testICMPChecksum(t, h.Checksum, func() uint16 {
		return header.ICMPv4Checksum(h, vv)
	}, want, fmt.Sprintf("header: {%v} data {%v}", h, vv))
}

func TestICMPv6Checksum(t *testing.T) {
	rnd := rand.New(rand.NewSource(42))

	h := header.ICMPv6(make([]byte, header.ICMPv6MinimumSize))
	if _, err := rnd.Read(h); err != nil {
		t.Fatalf("rnd.Read failed: %v", err)
	}
	h.SetChecksum(0)

	buf := make([]byte, 13)
	if _, err := rnd.Read(buf); err != nil {
		t.Fatalf("rnd.Read failed: %v", err)
	}
	vv := buffer.NewVectorisedView(len(buf), []buffer.View{
		buffer.NewViewFromBytes(buf[:7]),
		buffer.NewViewFromBytes(buf[7:10]),
		buffer.NewViewFromBytes(buf[10:]),
	})

	dst := header.IPv6Loopback
	src := header.IPv6Loopback

	want := header.PseudoHeaderChecksum(header.ICMPv6ProtocolNumber, src, dst, uint16(len(h)+vv.Size()))
	want = header.Checksum(vv.ToView(), want)
	want = ^header.Checksum(h, want)
	h.SetChecksum(want)

	testICMPChecksum(t, h.Checksum, func() uint16 {
		return header.ICMPv6Checksum(h, src, dst, vv)
	}, want, fmt.Sprintf("header: {%v} data {%v}", h, vv))
}
