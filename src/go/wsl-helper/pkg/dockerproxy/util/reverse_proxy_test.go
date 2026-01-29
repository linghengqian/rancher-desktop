/*
Copyright © 2021 SUSE LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFlushedWriterPeriodicFlush(t *testing.T) {
	t.Parallel()

	flusher := newRecordingFlusher()
	writer := newFlushedWriter(t.Context(), flusher)

	_, err := writer.Write([]byte("data"))
	assert.NoError(t, err)

	select {
	case <-flusher.flushCh:
	case <-time.After(flushInterval * flushTestTimeoutMultiplier):
		t.Fatal("expected periodic flush")
	}
}

func TestFlushedWriterStopFlushingSuppressesPeriodicFlush(t *testing.T) {
	t.Parallel()

	flusher := newRecordingFlusher()
	writer := newFlushedWriter(t.Context(), flusher)

	_, err := writer.Write([]byte("data"))
	assert.NoError(t, err)

	writer.stopFlushing()

	select {
	case <-flusher.flushCh:
		t.Fatal("unexpected flush after stopFlushing")
	case <-time.After(flushInterval * flushTestTimeoutMultiplier):
	}
}

const flushTestTimeoutMultiplier = 5

type recordingFlusher struct {
	io.Writer
	flushCh chan struct{}
}

func newRecordingFlusher() *recordingFlusher {
	return &recordingFlusher{
		Writer:  io.Discard,
		flushCh: make(chan struct{}, 1),
	}
}

func (writer *recordingFlusher) Flush() {
	close(writer.flushCh)
}
