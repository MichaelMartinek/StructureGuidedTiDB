// Copyright 2023 PingCAP, Inc.
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

// This was added by MichaelMartinek

package executor

import (
	"context"
	"github.com/pingcap/tidb/util/logutil"

	"github.com/pingcap/errors"
	"github.com/pingcap/tidb/util/chunk"
	"github.com/pingcap/tidb/util/disk"
	"github.com/pingcap/tidb/util/memory"
)

var _ Executor = &CTEMockExec{}

// CTEMockExec implements CTE, if there are no other CTEs (or CTETableReaders, not implemented, since not used) present
// It just takes the seedExec and hands the returned chunks along to the caller, without storing the tuples in the
// meantime and without using storage for this
// Meant to be used with Yannakakis algorithm query evaluation
type CTEMockExec struct {
	baseExecutor

	seedExec Executor

	emptyResult bool // if the result of the underlying operator is empty, this is set to true; used internally by the operator

	memTracker  *memory.Tracker
	diskTracker *disk.Tracker
}

// Open implements the Executor interface.
func (e *CTEMockExec) Open(ctx context.Context) (err error) {
	if err := e.baseExecutor.Open(ctx); err != nil {
		return err
	}

	if e.seedExec == nil {
		return errors.New("seedExec for CTEMockExec is nil")
	}
	if err = e.seedExec.Open(ctx); err != nil {
		return err
	}

	if e.memTracker != nil {
		e.memTracker.Reset()
	} else {
		e.memTracker = memory.NewTracker(e.id, -1)
	}
	e.diskTracker = disk.NewTracker(e.id, -1)
	e.memTracker.AttachTo(e.ctx.GetSessionVars().StmtCtx.MemTracker)
	e.diskTracker.AttachTo(e.ctx.GetSessionVars().StmtCtx.DiskTracker)

	// init empty
	e.emptyResult = true

	return nil
}

// Next implements the Executor interface.
func (e *CTEMockExec) Next(ctx context.Context, req *chunk.Chunk) (err error) {
	req.Reset()

	if err = Next(ctx, e.seedExec, req); err != nil {
		return err
	}

	// YAN early stop check: if the result was previously empty, and the current result is empty, then do a early stop
	if e.ctx.GetSessionVars().StmtCtx.GetIsYannakakis() {
		if e.emptyResult && req.NumRows() == 0 {
			e.ctx.GetSessionVars().StmtCtx.SetIsEmpty()
			logutil.BgLogger().Info("YAN early stop CTEMockExec")
		}
		e.emptyResult = false
	}
	// YAN early stop end

	return nil
}

// Close implements the Executor interface.
func (e *CTEMockExec) Close() (err error) {
	if err = e.seedExec.Close(); err != nil {
		return err
	}

	return e.baseExecutor.Close()
}
