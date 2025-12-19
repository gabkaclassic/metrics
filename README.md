File: main
Build ID: 88f15b23364780b56d65c61212e93ef5ff08776d
Type: cpu
Time: 2025-12-19 16:48:01 MSK
Duration: 120.08s, Total samples = 5.37s ( 4.47%)
Showing nodes accounting for -1.40s, 26.07% of 5.37s total
Dropped 17 nodes (cum <= 0.03s)
      flat  flat%   sum%        cum   cum%
    -0.21s  3.91%  3.91%     -0.21s  3.91%  crypto/internal/fips140/sha256.block
    -0.15s  2.79%  6.70%     -0.15s  2.79%  runtime.futex
    -0.11s  2.05%  8.75%     -0.44s  8.19%  runtime.scanobject
    -0.07s  1.30% 10.06%     -0.10s  1.86%  runtime.typePointers.next
    -0.06s  1.12% 11.17%     -0.06s  1.12%  runtime.memclrNoHeapPointers
    -0.06s  1.12% 12.29%     -0.07s  1.30%  runtime.step
     0.05s  0.93% 11.36%      0.04s  0.74%  runtime.findfunc
    -0.04s  0.74% 12.10%     -0.26s  4.84%  crypto/internal/fips140/sha256.(*Digest).Sum
    -0.04s  0.74% 12.85%     -0.03s  0.56%  internal/runtime/atomic.(*Uint64).Load (inline)
    -0.04s  0.74% 13.59%     -0.04s  0.74%  internal/runtime/atomic.Loadp
    -0.04s  0.74% 14.34%     -0.04s  0.74%  internal/runtime/atomic.Xchg8
    -0.04s  0.74% 15.08%     -0.05s  0.93%  runtime.(*gcBits).bitp (inline)
    -0.04s  0.74% 15.83%     -0.06s  1.12%  runtime.(*mspan).typePointersOfUnchecked
    -0.04s  0.74% 16.57%     -0.05s  0.93%  runtime.findObject
    -0.04s  0.74% 17.32%     -0.04s  0.74%  runtime.memmove
     0.03s  0.56% 16.76%     -0.30s  5.59%  github.com/lib/pq/scram.(*Client).saltPassword
     0.03s  0.56% 16.20%      0.03s  0.56%  internal/runtime/syscall.Syscall6
    -0.03s  0.56% 16.76%     -0.07s  1.30%  runtime.(*unwinder).resolveInternal
     0.03s  0.56% 16.20%      0.03s  0.56%  runtime.cgocall
    -0.03s  0.56% 16.76%     -0.05s  0.93%  runtime.getGCMask (inline)
    -0.03s  0.56% 17.32%     -0.08s  1.49%  runtime.makeslice
    -0.03s  0.56% 17.88%     -0.03s  0.56%  runtime.nanotime (inline)
    -0.03s  0.56% 18.44%      0.03s  0.56%  runtime.suspendG
    -0.03s  0.56% 18.99%     -0.03s  0.56%  runtime.usleep
    -0.02s  0.37% 19.37%     -0.01s  0.19%  crypto/internal/fips140/sha256.consumeUint32 (inline)
     0.02s  0.37% 18.99%      0.02s  0.37%  internal/runtime/atomic.Load64
    -0.02s  0.37% 19.37%     -0.02s  0.37%  internal/runtime/atomic.Xadd
    -0.02s  0.37% 19.74%     -0.02s  0.37%  memeqbody
    -0.02s  0.37% 20.11%     -0.02s  0.37%  runtime.(*spanSet).reset
    -0.02s  0.37% 20.48%     -0.03s  0.56%  runtime.(*stkframe).getStackMap
    -0.02s  0.37% 20.86%     -0.06s  1.12%  runtime.(*sweepLocked).sweep
    -0.02s  0.37% 21.23%     -0.02s  0.37%  runtime.arenaIndex (inline)
     0.02s  0.37% 20.86%      0.02s  0.37%  runtime.duffcopy
     0.02s  0.37% 20.48%      0.02s  0.37%  runtime.getempty
    -0.02s  0.37% 20.86%     -0.04s  0.74%  runtime.mapaccess1_faststr
     0.02s  0.37% 20.48%      0.02s  0.37%  runtime.osyield
     0.02s  0.37% 20.11%      0.01s  0.19%  runtime.pMask.read
    -0.02s  0.37% 20.48%     -0.02s  0.37%  runtime.procyield
    -0.02s  0.37% 20.86%     -0.04s  0.74%  runtime.scanframeworker
    -0.02s  0.37% 21.23%     -0.02s  0.37%  syscall.socketcall
    -0.01s  0.19% 21.42%     -0.03s  0.56%  bufio.(*Writer).Flush
     0.01s  0.19% 21.23%     -0.01s  0.19%  bytes.(*Buffer).WriteByte
     0.01s  0.19% 21.04%      0.01s  0.19%  compress/flate.(*decompressor).makeReader
    -0.01s  0.19% 21.23%     -0.01s  0.19%  compress/flate.(*decompressor).moreBits
     0.01s  0.19% 21.04%      0.01s  0.19%  context.WithValue
    -0.01s  0.19% 21.23%     -0.01s  0.19%  crypto/internal/fips140/hmac.(*HMAC).Reset
    -0.01s  0.19% 21.42%     -0.02s  0.37%  crypto/internal/fips140/hmac.(*HMAC).Write
    -0.01s  0.19% 21.60%     -0.02s  0.37%  crypto/internal/fips140/sha256.(*Digest).UnmarshalBinary
    -0.01s  0.19% 21.79%     -0.21s  3.91%  crypto/internal/fips140/sha256.(*Digest).checkSum
     0.01s  0.19% 21.60%      0.01s  0.19%  crypto/internal/fips140deps/byteorder.BEUint32 (inline)
    -0.01s  0.19% 21.79%     -0.29s  5.40%  database/sql.(*DB).BeginTx
     0.01s  0.19% 21.60%      0.01s  0.19%  encoding/json.appendString[go.shape.string]
     0.01s  0.19% 21.42%      0.01s  0.19%  encoding/json.stateBeginValue
     0.01s  0.19% 21.23%     -0.03s  0.56%  github.com/gabkaclassic/metrics/internal/audit.urlHandler.handle
    -0.01s  0.19% 21.42%     -0.03s  0.56%  github.com/gabkaclassic/metrics/internal/service.(*metricsService).SaveAll
     0.01s  0.19% 21.23%      0.03s  0.56%  github.com/lib/pq.(*conn).recvMessage
     0.01s  0.19% 21.04%      0.01s  0.19%  github.com/lib/pq.(*writeBuf).int16 (inline)
     0.01s  0.19% 20.86%      0.01s  0.19%  gosave_systemstack_switch
     0.01s  0.19% 20.67%      0.01s  0.19%  hash/crc32.simpleUpdate (inline)
    -0.01s  0.19% 20.86%     -0.01s  0.19%  internal/abi.Name.Name
    -0.01s  0.19% 21.04%     -0.01s  0.19%  internal/bytealg.IndexByte
    -0.01s  0.19% 21.23%     -0.01s  0.19%  internal/byteorder.BEPutUint64 (inline)
    -0.01s  0.19% 21.42%     -0.01s  0.19%  internal/chacha8rand.qr (inline)
    -0.01s  0.19% 21.60%     -0.01s  0.19%  internal/poll.(*fdMutex).increfAndClose
     0.01s  0.19% 21.42%      0.03s  0.56%  internal/poll.(*fdMutex).rwlock
    -0.01s  0.19% 21.60%     -0.01s  0.19%  internal/poll.(*fdMutex).rwunlock
    -0.01s  0.19% 21.79%     -0.03s  0.56%  internal/poll.setDeadlineImpl
    -0.01s  0.19% 21.97%     -0.01s  0.19%  internal/runtime/atomic.(*Int64).Load (inline)
     0.01s  0.19% 21.79%      0.01s  0.19%  internal/runtime/atomic.(*Uint32).Store (inline)
     0.01s  0.19% 21.60%     -0.01s  0.19%  internal/runtime/atomic.(*UnsafePointer).Load (inline)
    -0.01s  0.19% 21.79%     -0.01s  0.19%  internal/runtime/atomic.Cas
     0.01s  0.19% 21.60%      0.01s  0.19%  internal/runtime/atomic.Cas64
     0.01s  0.19% 21.42%      0.01s  0.19%  internal/runtime/atomic.Load8
    -0.01s  0.19% 21.60%     -0.01s  0.19%  internal/runtime/atomic.Or8
     0.01s  0.19% 21.42%      0.01s  0.19%  internal/runtime/atomic.Store
    -0.01s  0.19% 21.60%     -0.01s  0.19%  internal/runtime/atomic.Xchg64
    -0.01s  0.19% 21.79%     -0.01s  0.19%  internal/runtime/maps.(*Iter).Next
     0.01s  0.19% 21.60%      0.01s  0.19%  internal/runtime/maps.(*Map).Delete
     0.01s  0.19% 21.42%      0.01s  0.19%  internal/runtime/maps.(*groupReference).ctrls (inline)
    -0.01s  0.19% 21.60%     -0.01s  0.19%  internal/runtime/maps.NewMap
     0.01s  0.19% 21.42%      0.01s  0.19%  internal/runtime/sys.Len64 (inline)
    -0.01s  0.19% 21.60%     -0.01s  0.19%  internal/runtime/sys.OnesCount64 (inline)
     0.01s  0.19% 21.42%      0.05s  0.93%  net.(*netFD).Read
    -0.01s  0.19% 21.60%     -0.01s  0.19%  net.stringsHasSuffixFold
    -0.01s  0.19% 21.79%     -0.09s  1.68%  net/http.(*Transport).dial
    -0.01s  0.19% 21.97%     -0.01s  0.19%  net/http.(*body).Close
     0.01s  0.19% 21.79%     -0.04s  0.74%  net/http.(*body).Read
    -0.01s  0.19% 21.97%     -0.04s  0.74%  net/http.(*body).readLocked
    -0.01s  0.19% 22.16%     -0.01s  0.19%  net/http.Header.sortedKeyValues
     0.01s  0.19% 21.97%      0.01s  0.19%  net/http.appendTime
    -0.01s  0.19% 22.16%     -0.03s  0.56%  net/textproto.canonicalMIMEHeaderKey
    -0.01s  0.19% 22.35%     -0.01s  0.19%  reflect.(*rtype).exportedMethods
     0.01s  0.19% 22.16%      0.01s  0.19%  reflect.Value.SetFloat
    -0.01s  0.19% 22.35%     -0.02s  0.37%  runtime.(*activeSweep).begin
    -0.01s  0.19% 22.53%     -0.01s  0.19%  runtime.(*bucket).mp
    -0.01s  0.19% 22.72%     -0.01s  0.19%  runtime.(*fixalloc).alloc
     0.01s  0.19% 22.53%      0.01s  0.19%  runtime.(*gQueue).pop (inline)
    -0.01s  0.19% 22.72%     -0.01s  0.19%  runtime.(*gcBits).bytep (inline)
    -0.01s  0.19% 22.91%     -0.03s  0.56%  runtime.(*gcBitsArena).tryAlloc
    -0.01s  0.19% 23.09%     -0.01s  0.19%  runtime.(*gcCPULimiterState).accumulate
    -0.01s  0.19% 23.28%     -0.01s  0.19%  runtime.(*gcControllerState).trigger
    -0.01s  0.19% 23.46%     -0.01s  0.19%  runtime.(*gcWork).dispose
     0.01s  0.19% 23.28%      0.01s  0.19%  runtime.(*gcWork).tryGetFast (inline)
    -0.01s  0.19% 23.46%     -0.01s  0.19%  runtime.(*itabTableType).find
     0.01s  0.19% 23.28%      0.01s  0.19%  runtime.(*m).snapshotAllp
     0.01s  0.19% 23.09%      0.01s  0.19%  runtime.(*mcentral).grow
    -0.01s  0.19% 23.28%     -0.05s  0.93%  runtime.(*mcentral).uncacheSpan
    -0.01s  0.19% 23.46%     -0.01s  0.19%  runtime.(*moduledata).textOff (inline)
    -0.01s  0.19% 23.65%     -0.01s  0.19%  runtime.(*mspan).heapBitsSmallForAddr
     0.01s  0.19% 23.46%      0.01s  0.19%  runtime.(*mspan).nextFreeIndex
    -0.01s  0.19% 23.65%     -0.01s  0.19%  runtime.(*mspan).refillAllocCache
    -0.01s  0.19% 23.84%     -0.01s  0.19%  runtime.(*pageAlloc).chunkOf (inline)
     0.01s  0.19% 23.65%      0.01s  0.19%  runtime.(*pageAlloc).find.func1
     0.01s  0.19% 23.46%      0.02s  0.37%  runtime.(*pageAlloc).free
    -0.01s  0.19% 23.65%     -0.03s  0.56%  runtime.(*pageAlloc).scavenge
    -0.01s  0.19% 23.84%     -0.01s  0.19%  runtime.(*pageBits).popcntRange
     0.01s  0.19% 23.65%      0.01s  0.19%  runtime.(*pageCache).alloc
    -0.01s  0.19% 23.84%     -0.01s  0.19%  runtime.(*pallocData).findScavengeCandidate
    -0.01s  0.19% 24.02%     -0.01s  0.19%  runtime.(*randomEnum).next (inline)
    -0.01s  0.19% 24.21%     -0.01s  0.19%  runtime.(*randomOrder).start (inline)
    -0.01s  0.19% 24.39%     -0.01s  0.19%  runtime.(*spanSet).push
     0.01s  0.19% 24.21%      0.01s  0.19%  runtime.(*sweepLocker).tryAcquire
    -0.01s  0.19% 24.39%     -0.02s  0.37%  runtime.(*timers).wakeTime
    -0.01s  0.19% 24.58%     -0.01s  0.19%  runtime.(*waitq).dequeue
     0.01s  0.19% 24.39%      0.01s  0.19%  runtime._d2v
    -0.01s  0.19% 24.58%     -0.01s  0.19%  runtime._div64by32
     0.01s  0.19% 24.39%      0.01s  0.19%  runtime.acquireSudog
     0.01s  0.19% 24.21%      0.01s  0.19%  runtime.atomicwb
     0.01s  0.19% 24.02%     -0.05s  0.93%  runtime.bgsweep
     0.01s  0.19% 23.84%     -0.01s  0.19%  runtime.casgstatus
    -0.01s  0.19% 24.02%     -0.01s  0.19%  runtime.cgoCheckUnknownPointer
    -0.01s  0.19% 24.21%     -0.01s  0.19%  runtime.chanrecv
    -0.01s  0.19% 24.39%     -0.01s  0.19%  runtime.cheaprand (inline)
     0.01s  0.19% 24.21%     -0.02s  0.37%  runtime.copystack
    -0.01s  0.19% 24.39%     -0.23s  4.28%  runtime.deductAssistCredit
     0.01s  0.19% 24.21%      0.01s  0.19%  runtime.duffzero
    -0.01s  0.19% 24.39%     -0.01s  0.19%  runtime.funcInfo.entry (inline)
    -0.01s  0.19% 24.58%     -0.02s  0.37%  runtime.funcMaxSPDelta
     0.01s  0.19% 24.39%     -0.04s  0.74%  runtime.funcspdelta (inline)
    -0.01s  0.19% 24.58%     -0.29s  5.40%  runtime.gcDrain
    -0.01s  0.19% 24.77%     -0.23s  4.28%  runtime.gcDrainN
     0.01s  0.19% 24.58%     -0.08s  1.49%  runtime.gcMarkDone
    -0.01s  0.19% 24.77%     -0.01s  0.19%  runtime.gcResetMarkState.func1
     0.01s  0.19% 24.58%      0.01s  0.19%  runtime.gclinkptr.ptr (inline)
    -0.01s  0.19% 24.77%     -0.01s  0.19%  runtime.getpid
    -0.01s  0.19% 24.95%     -0.07s  1.30%  runtime.greyobject
    -0.01s  0.19% 25.14%     -0.02s  0.37%  runtime.growslice
    -0.01s  0.19% 25.33%     -0.01s  0.19%  runtime.isSystemGoroutine
    -0.01s  0.19% 25.51%     -0.09s  1.68%  runtime.lock2
     0.01s  0.19% 25.33%     -0.08s  1.49%  runtime.lockWithRank (inline)
    -0.01s  0.19% 25.51%     -0.01s  0.19%  runtime.mProf_Free
     0.01s  0.19% 25.33%      0.01s  0.19%  runtime.madvise
    -0.01s  0.19% 25.51%     -0.04s  0.74%  runtime.mallocgcSmallScanNoHeader
    -0.01s  0.19% 25.70%     -0.05s  0.93%  runtime.markroot.func1
    -0.01s  0.19% 25.88%     -0.01s  0.19%  runtime.mergeSummaries
    -0.01s  0.19% 26.07%     -0.01s  0.19%  runtime.mget (inline)
     0.01s  0.19% 25.88%     -0.07s  1.30%  runtime.pcvalue
    -0.01s  0.19% 26.07%     -0.01s  0.19%  runtime.readUintptr (inline)
    -0.01s  0.19% 26.26%     -0.01s  0.19%  runtime.readvarint (inline)
     0.01s  0.19% 26.07%      0.01s  0.19%  runtime.ready
    -0.01s  0.19% 26.26%     -0.01s  0.19%  runtime.releaseSudog
    -0.01s  0.19% 26.44%     -0.01s  0.19%  runtime.releasem (inline)
     0.01s  0.19% 26.26%     -0.02s  0.37%  runtime.runqgrab
     0.01s  0.19% 26.07%      0.02s  0.37%  runtime.selunlock
     0.01s  0.19% 25.88%      0.01s  0.19%  runtime.shrinkstack
    -0.01s  0.19% 26.07%     -0.01s  0.19%  runtime.spanClass.sizeclass (inline)
    -0.01s  0.19% 26.26%     -0.01s  0.19%  runtime.spanOfHeap
    -0.01s  0.19% 26.44%     -0.01s  0.19%  runtime.spanOfUnchecked (inline)
    -0.01s  0.19% 26.63%     -0.06s  1.12%  runtime.stealWork
    -0.01s  0.19% 26.82%     -0.01s  0.19%  runtime.strhash
     0.01s  0.19% 26.63%      0.02s  0.37%  runtime.tracebackPCs
    -0.01s  0.19% 26.82%     -0.01s  0.19%  runtime.typePointers.fastForward
     0.01s  0.19% 26.63%      0.01s  0.19%  runtime.uint32tofloat64
     0.01s  0.19% 26.44%      0.01s  0.19%  runtime.uint64div
     0.01s  0.19% 26.26%     -0.01s  0.19%  runtime.unlock2
     0.01s  0.19% 26.07%      0.01s  0.19%  strconv.atof64exact
     0.01s  0.19% 25.88%      0.02s  0.37%  sync.(*Pool).pin
    -0.01s  0.19% 26.07%     -0.01s  0.19%  sync/atomic.(*Pointer[go.shape.struct { sync.poolDequeue; sync.next sync/atomic.Pointer[sync.poolChainElt]; sync.prev sync/atomic.Pointer[sync.poolChainElt] }]).Load (inline)
    -0.01s  0.19% 26.26%     -0.01s  0.19%  sync/atomic.(*Value).Load
     0.01s  0.19% 26.07%      0.01s  0.19%  time.runtimeNow
     0.01s  0.19% 25.88%      0.01s  0.19%  unicode/utf8.DecodeRuneInString
    -0.01s  0.19% 26.07%     -0.01s  0.19%  vendor/golang.org/x/net/http/httpguts.ValidHeaderFieldName (inline)
         0     0% 26.07%      0.02s  0.37%  bufio.(*Reader).Peek
         0     0% 26.07%      0.01s  0.19%  bufio.(*Reader).Read
         0     0% 26.07%      0.01s  0.19%  bufio.(*Reader).ReadLine
         0     0% 26.07%      0.01s  0.19%  bufio.(*Reader).ReadSlice
         0     0% 26.07%      0.03s  0.56%  bufio.(*Reader).fill
         0     0% 26.07%      0.01s  0.19%  bufio.(*Reader).reset (inline)
         0     0% 26.07%     -0.01s  0.19%  bufio.(*Writer).ReadFrom
         0     0% 26.07%      0.01s  0.19%  bufio.NewReader (inline)
         0     0% 26.07%     -0.01s  0.19%  bytes.(*Buffer).Write
         0     0% 26.07%     -0.03s  0.56%  bytes.(*Buffer).grow
         0     0% 26.07%     -0.01s  0.19%  bytes.(*Reader).Read
         0     0% 26.07%     -0.01s  0.19%  bytes.IndexByte (inline)
         0     0% 26.07%     -0.01s  0.19%  bytes.growSlice
         0     0% 26.07%     -0.08s  1.49%  compress/flate.(*compressor).init
         0     0% 26.07%     -0.01s  0.19%  compress/flate.(*decompressor).Read
         0     0% 26.07%     -0.01s  0.19%  compress/flate.(*decompressor).nextBlock
         0     0% 26.07%     -0.01s  0.19%  compress/flate.(*dictDecoder).init (inline)
         0     0% 26.07%      0.01s  0.19%  compress/flate.NewReader
         0     0% 26.07%     -0.10s  1.86%  compress/flate.NewWriter (inline)
         0     0% 26.07%     -0.06s  1.12%  compress/flate.newDeflateFast (inline)
         0     0% 26.07%     -0.02s  0.37%  compress/flate.newHuffmanBitWriter (inline)
         0     0% 26.07%     -0.01s  0.19%  compress/flate.newHuffmanEncoder (inline)
         0     0% 26.07%      0.01s  0.19%  compress/gzip.(*Reader).Reset
         0     0% 26.07%      0.01s  0.19%  compress/gzip.(*Reader).readHeader
         0     0% 26.07%     -0.10s  1.86%  compress/gzip.(*Writer).Close
         0     0% 26.07%     -0.10s  1.86%  compress/gzip.(*Writer).Write
         0     0% 26.07%      0.01s  0.19%  compress/gzip.NewReader (inline)
         0     0% 26.07%     -0.01s  0.19%  context.(*cancelCtx).Done
         0     0% 26.07%     -0.28s  5.21%  crypto/internal/fips140/hmac.(*HMAC).Sum
         0     0% 26.07%     -0.22s  4.10%  crypto/internal/fips140/sha256.(*Digest).Write
         0     0% 26.07%     -0.01s  0.19%  crypto/internal/fips140deps/byteorder.BEPutUint64 (inline)
         0     0% 26.07%     -0.02s  0.37%  crypto/internal/sysrand.Read
         0     0% 26.07%     -0.02s  0.37%  crypto/internal/sysrand.read
         0     0% 26.07%     -0.02s  0.37%  crypto/rand.(*reader).Read
         0     0% 26.07%     -0.01s  0.19%  crypto/rand.Read
         0     0% 26.07%     -0.28s  5.21%  database/sql.(*DB).BeginTx.func1
         0     0% 26.07%     -0.28s  5.21%  database/sql.(*DB).begin
         0     0% 26.07%      0.01s  0.19%  database/sql.(*DB).beginDC
         0     0% 26.07%      0.01s  0.19%  database/sql.(*DB).beginDC.func1
         0     0% 26.07%     -0.29s  5.40%  database/sql.(*DB).conn
         0     0% 26.07%      0.02s  0.37%  database/sql.(*DB).execDC
         0     0% 26.07%      0.02s  0.37%  database/sql.(*DB).execDC.func2
         0     0% 26.07%     -0.28s  5.21%  database/sql.(*DB).retry
         0     0% 26.07%      0.01s  0.19%  database/sql.(*Tx).Commit.func1
         0     0% 26.07%      0.02s  0.37%  database/sql.(*Tx).ExecContext
         0     0% 26.07%      0.01s  0.19%  database/sql.(*driverConn).resetSession
         0     0% 26.07%      0.01s  0.19%  database/sql.ctxDriverBegin
         0     0% 26.07%      0.03s  0.56%  database/sql.ctxDriverExec
         0     0% 26.07%     -0.01s  0.19%  database/sql.defaultCheckNamedValue
         0     0% 26.07%     -0.01s  0.19%  database/sql.driverArgsConnLocked
         0     0% 26.07%     -0.29s  5.40%  database/sql.dsnConnector.Connect
         0     0% 26.07%      0.04s  0.74%  database/sql.withLock
         0     0% 26.07%     -0.01s  0.19%  database/sql/driver.callValuerValue
         0     0% 26.07%     -0.01s  0.19%  database/sql/driver.defaultConverter.ConvertValue
         0     0% 26.07%     -0.02s  0.37%  encoding/json.(*Decoder).Decode
         0     0% 26.07%      0.02s  0.37%  encoding/json.(*Encoder).Encode
         0     0% 26.07%     -0.02s  0.37%  encoding/json.(*decodeState).array
         0     0% 26.07%      0.01s  0.19%  encoding/json.(*decodeState).literalStore
         0     0% 26.07%      0.01s  0.19%  encoding/json.(*decodeState).scanWhile
         0     0% 26.07%     -0.02s  0.37%  encoding/json.(*decodeState).unmarshal
         0     0% 26.07%     -0.02s  0.37%  encoding/json.(*decodeState).value
         0     0% 26.07%     -0.02s  0.37%  encoding/json.(*encodeState).marshal
         0     0% 26.07%     -0.02s  0.37%  encoding/json.(*encodeState).reflectValue
         0     0% 26.07%     -0.05s  0.93%  encoding/json.Marshal
         0     0% 26.07%      0.01s  0.19%  encoding/json.arrayEncoder.encode
         0     0% 26.07%     -0.02s  0.37%  encoding/json.indirect
         0     0% 26.07%      0.02s  0.37%  encoding/json.mapEncoder.encode
         0     0% 26.07%     -0.02s  0.37%  encoding/json.newEncodeState
         0     0% 26.07%     -0.04s  0.74%  encoding/json.ptrEncoder.encode
         0     0% 26.07%      0.01s  0.19%  encoding/json.sliceEncoder.encode
         0     0% 26.07%     -0.03s  0.56%  encoding/json.structEncoder.encode
         0     0% 26.07%     -0.01s  0.19%  encoding/json.typeEncoder
         0     0% 26.07%     -0.01s  0.19%  encoding/json.valueEncoder
         0     0% 26.07%     -0.01s  0.19%  github.com/gabkaclassic/metrics/internal/audit.(*auditor).AuditMany.func1
         0     0% 26.07%      0.02s  0.37%  github.com/gabkaclassic/metrics/internal/audit.fileHandler.handle
         0     0% 26.07%      0.01s  0.19%  github.com/gabkaclassic/metrics/internal/audit.getMetricsNames
         0     0% 26.07%     -0.05s  0.93%  github.com/gabkaclassic/metrics/internal/handler.(*MetricsHandler).SaveAll
         0     0% 26.07%     -0.14s  2.61%  github.com/gabkaclassic/metrics/internal/handler.SetupRouter.Decompress.func2.1
         0     0% 26.07%     -0.14s  2.61%  github.com/gabkaclassic/metrics/internal/handler.SetupRouter.SignVerify.func3.1
         0     0% 26.07%     -0.14s  2.61%  github.com/gabkaclassic/metrics/internal/handler.setupMetricsRouter.Compress.func7.1
         0     0% 26.07%     -0.05s  0.93%  github.com/gabkaclassic/metrics/internal/handler.setupMetricsRouter.RequireContentType.func6.1
         0     0% 26.07%     -0.14s  2.61%  github.com/gabkaclassic/metrics/internal/handler.setupMetricsRouter.WithContentType.func8.1
         0     0% 26.07%     -0.20s  3.72%  github.com/gabkaclassic/metrics/internal/repository.(*dbMetricsRepository).AddAll
         0     0% 26.07%     -0.20s  3.72%  github.com/gabkaclassic/metrics/internal/repository.(*dbMetricsRepository).AddAll.func1
         0     0% 26.07%     -0.07s  1.30%  github.com/gabkaclassic/metrics/internal/repository.(*dbMetricsRepository).ResetAll
         0     0% 26.07%     -0.07s  1.30%  github.com/gabkaclassic/metrics/internal/repository.(*dbMetricsRepository).ResetAll.func1
         0     0% 26.07%     -0.27s  5.03%  github.com/gabkaclassic/metrics/internal/repository.(*dbMetricsRepository).executeWithRetry
         0     0% 26.07%     -0.20s  3.72%  github.com/gabkaclassic/metrics/internal/service.(*metricsService).SaveAll.func1
         0     0% 26.07%     -0.07s  1.30%  github.com/gabkaclassic/metrics/internal/service.(*metricsService).SaveAll.func3
         0     0% 26.07%     -0.10s  1.86%  github.com/gabkaclassic/metrics/pkg/compress.(*CompressWriter).Close
         0     0% 26.07%      0.01s  0.19%  github.com/gabkaclassic/metrics/pkg/compress.NewGzipWriter
         0     0% 26.07%     -0.01s  0.19%  github.com/gabkaclassic/metrics/pkg/httpclient.(*Client).Post
         0     0% 26.07%     -0.01s  0.19%  github.com/gabkaclassic/metrics/pkg/httpclient.(*Client).do
         0     0% 26.07%     -0.14s  2.61%  github.com/gabkaclassic/metrics/pkg/middleware.AuditContext.func1
         0     0% 26.07%     -0.21s  3.91%  github.com/gabkaclassic/metrics/pkg/middleware.Logger.func1
         0     0% 26.07%     -0.22s  4.10%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 26.07%     -0.14s  2.61%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 26.07%     -0.01s  0.19%  github.com/google/uuid.NewRandomFromReader
         0     0% 26.07%     -0.24s  4.47%  github.com/lib/pq.(*Connector).open
         0     0% 26.07%      0.01s  0.19%  github.com/lib/pq.(*conn).BeginTx
         0     0% 26.07%      0.01s  0.19%  github.com/lib/pq.(*conn).Commit
         0     0% 26.07%      0.03s  0.56%  github.com/lib/pq.(*conn).Exec
         0     0% 26.07%      0.03s  0.56%  github.com/lib/pq.(*conn).ExecContext
         0     0% 26.07%     -0.28s  5.21%  github.com/lib/pq.(*conn).auth
         0     0% 26.07%      0.01s  0.19%  github.com/lib/pq.(*conn).begin
         0     0% 26.07%     -0.01s  0.19%  github.com/lib/pq.(*conn).parseComplete
         0     0% 26.07%      0.05s  0.93%  github.com/lib/pq.(*conn).prepareTo
         0     0% 26.07%     -0.01s  0.19%  github.com/lib/pq.(*conn).readExecuteResponse
         0     0% 26.07%      0.03s  0.56%  github.com/lib/pq.(*conn).readParseResponse
         0     0% 26.07%      0.01s  0.19%  github.com/lib/pq.(*conn).readStatementDescribeResponse
         0     0% 26.07%      0.01s  0.19%  github.com/lib/pq.(*conn).recv
         0     0% 26.07%      0.03s  0.56%  github.com/lib/pq.(*conn).recv1 (inline)
         0     0% 26.07%      0.03s  0.56%  github.com/lib/pq.(*conn).recv1Buf
         0     0% 26.07%      0.02s  0.37%  github.com/lib/pq.(*conn).send
         0     0% 26.07%      0.02s  0.37%  github.com/lib/pq.(*conn).simpleExec
         0     0% 26.07%     -0.29s  5.40%  github.com/lib/pq.(*conn).startup
         0     0% 26.07%      0.02s  0.37%  github.com/lib/pq.(*conn).watchCancel.func1
         0     0% 26.07%     -0.02s  0.37%  github.com/lib/pq.(*readBuf).string
         0     0% 26.07%     -0.02s  0.37%  github.com/lib/pq.(*stmt).Exec
         0     0% 26.07%     -0.01s  0.19%  github.com/lib/pq.(*stmt).exec
         0     0% 26.07%     -0.29s  5.40%  github.com/lib/pq.DialOpen
         0     0% 26.07%     -0.29s  5.40%  github.com/lib/pq.Driver.Open
         0     0% 26.07%     -0.01s  0.19%  github.com/lib/pq.Float64Array.Value
         0     0% 26.07%     -0.05s  0.93%  github.com/lib/pq.NewConnector
         0     0% 26.07%     -0.29s  5.40%  github.com/lib/pq.Open (inline)
         0     0% 26.07%     -0.03s  0.56%  github.com/lib/pq.ParseURL
         0     0% 26.07%      0.05s  0.93%  github.com/lib/pq.defaultDialer.DialContext
         0     0% 26.07%      0.05s  0.93%  github.com/lib/pq.dial
         0     0% 26.07%     -0.01s  0.19%  github.com/lib/pq.parseEnviron
         0     0% 26.07%     -0.30s  5.59%  github.com/lib/pq/scram.(*Client).Step
         0     0% 26.07%      0.01s  0.19%  github.com/lib/pq/scram.(*Client).clientProof
         0     0% 26.07%     -0.01s  0.19%  github.com/lib/pq/scram.(*Client).step1
         0     0% 26.07%     -0.29s  5.40%  github.com/lib/pq/scram.(*Client).step2
         0     0% 26.07%      0.01s  0.19%  github.com/lib/pq/scram.NewClient
         0     0% 26.07%      0.01s  0.19%  hash/crc32.Update (inline)
         0     0% 26.07%      0.01s  0.19%  hash/crc32.init.func2.1
         0     0% 26.07%      0.01s  0.19%  hash/crc32.slicingUpdate
         0     0% 26.07%      0.01s  0.19%  hash/crc32.update
         0     0% 26.07%     -0.02s  0.37%  internal/bytealg.MakeNoZero
         0     0% 26.07%     -0.01s  0.19%  internal/chacha8rand.(*State).Refill
         0     0% 26.07%     -0.01s  0.19%  internal/chacha8rand.block_generic
         0     0% 26.07%      0.04s  0.74%  internal/poll.(*FD).Close
         0     0% 26.07%      0.03s  0.56%  internal/poll.(*FD).Fsync
         0     0% 26.07%      0.03s  0.56%  internal/poll.(*FD).Fsync.func1 (inline)
         0     0% 26.07%     -0.01s  0.19%  internal/poll.(*FD).Init
         0     0% 26.07%      0.04s  0.74%  internal/poll.(*FD).Read
         0     0% 26.07%     -0.02s  0.37%  internal/poll.(*FD).SetReadDeadline (inline)
         0     0% 26.07%     -0.01s  0.19%  internal/poll.(*FD).SetWriteDeadline (inline)
         0     0% 26.07%      0.03s  0.56%  internal/poll.(*FD).SetsockoptInt
         0     0% 26.07%      0.03s  0.56%  internal/poll.(*FD).Write
         0     0% 26.07%      0.05s  0.93%  internal/poll.(*FD).decref
         0     0% 26.07%      0.05s  0.93%  internal/poll.(*FD).destroy
         0     0% 26.07%     -0.01s  0.19%  internal/poll.(*FD).readUnlock
         0     0% 26.07%      0.03s  0.56%  internal/poll.(*FD).writeLock (inline)
         0     0% 26.07%      0.04s  0.74%  internal/poll.(*SysFile).destroy (inline)
         0     0% 26.07%      0.01s  0.19%  internal/poll.(*pollDesc).close (inline)
         0     0% 26.07%     -0.01s  0.19%  internal/poll.(*pollDesc).init
         0     0% 26.07%      0.03s  0.56%  internal/poll.ignoringEINTR (inline)
         0     0% 26.07%      0.05s  0.93%  internal/poll.ignoringEINTRIO (inline)
         0     0% 26.07%      0.01s  0.19%  internal/poll.runtime_pollClose
         0     0% 26.07%     -0.01s  0.19%  internal/poll.runtime_pollOpen
         0     0% 26.07%     -0.01s  0.19%  internal/runtime/atomic.(*Bool).Load (inline)
         0     0% 26.07%     -0.01s  0.19%  internal/runtime/atomic.(*Int32).Add (inline)
         0     0% 26.07%      0.01s  0.19%  internal/runtime/atomic.(*Int64).Add (inline)
         0     0% 26.07%     -0.01s  0.19%  internal/runtime/atomic.(*Int64).Swap (inline)
         0     0% 26.07%     -0.02s  0.37%  internal/runtime/atomic.(*Pointer[go.shape.struct { runtime.lfnode; runtime.popped internal/runtime/atomic.Uint32; runtime.spans [512]runtime.atomicMSpanPointer }]).Load (inline)
         0     0% 26.07%      0.02s  0.37%  internal/runtime/atomic.(*Uint32).Add (inline)
         0     0% 26.07%      0.01s  0.19%  internal/runtime/atomic.(*Uint64).CompareAndSwap (inline)
         0     0% 26.07%      0.01s  0.19%  internal/runtime/atomic.(*Uintptr).Store (inline)
         0     0% 26.07%     -0.02s  0.37%  internal/runtime/maps.(*Map).getWithoutKeySmallFastStr
         0     0% 26.07%     -0.01s  0.19%  internal/runtime/maps.(*Map).growToSmall
         0     0% 26.07%      0.01s  0.19%  internal/runtime/maps.(*Map).putSlotSmallFastStr
         0     0% 26.07%     -0.03s  0.56%  internal/runtime/maps.NewEmptyMap (inline)
         0     0% 26.07%     -0.01s  0.19%  internal/runtime/maps.newGroups (inline)
         0     0% 26.07%     -0.01s  0.19%  internal/runtime/maps.newarray
         0     0% 26.07%     -0.01s  0.19%  internal/runtime/maps.rand
         0     0% 26.07%      0.01s  0.19%  internal/runtime/sys.LeadingZeros64 (inline)
         0     0% 26.07%     -0.06s  1.12%  internal/runtime/syscall.EpollWait
         0     0% 26.07%     -0.03s  0.56%  internal/singleflight.(*Group).doCall
         0     0% 26.07%     -0.01s  0.19%  internal/stringslite.HasPrefix (inline)
         0     0% 26.07%     -0.01s  0.19%  internal/sync.(*HashTrieMap[go.shape.interface {},go.shape.interface {}]).Load
         0     0% 26.07%     -0.01s  0.19%  internal/sync.(*HashTrieMap[go.shape.interface {},go.shape.interface {}]).init
         0     0% 26.07%      0.01s  0.19%  internal/sync.(*Mutex).lockSlow
         0     0% 26.07%      0.01s  0.19%  internal/sync.runtime_doSpin
         0     0% 26.07%     -0.02s  0.37%  internal/syscall/unix.GetRandom
         0     0% 26.07%      0.01s  0.19%  io.Copy (inline)
         0     0% 26.07%     -0.01s  0.19%  io.CopyBuffer
         0     0% 26.07%      0.01s  0.19%  io.CopyN
         0     0% 26.07%     -0.06s  1.12%  io.ReadAll
         0     0% 26.07%      0.01s  0.19%  io.ReadAtLeast
         0     0% 26.07%      0.01s  0.19%  io.ReadFull (inline)
         0     0% 26.07%      0.01s  0.19%  io.discard.ReadFrom
         0     0% 26.07%      0.01s  0.19%  log/slog.(*JSONHandler).Handle
         0     0% 26.07%     -0.01s  0.19%  log/slog.(*Logger).log
         0     0% 26.07%      0.01s  0.19%  log/slog.(*commonHandler).handle
         0     0% 26.07%     -0.01s  0.19%  log/slog.(*commonHandler).newHandleState
         0     0% 26.07%      0.03s  0.56%  log/slog.(*handleState).appendAttr
         0     0% 26.07%     -0.01s  0.19%  log/slog.(*handleState).appendKey
         0     0% 26.07%      0.03s  0.56%  log/slog.(*handleState).appendNonBuiltIns
         0     0% 26.07%      0.03s  0.56%  log/slog.(*handleState).appendNonBuiltIns.func1 (inline)
         0     0% 26.07%      0.01s  0.19%  log/slog.(*handleState).appendTime
         0     0% 26.07%      0.03s  0.56%  log/slog.(*handleState).appendValue
         0     0% 26.07%     -0.02s  0.37%  log/slog.(*handleState).free
         0     0% 26.07%     -0.01s  0.19%  log/slog.Debug
         0     0% 26.07%      0.03s  0.56%  log/slog.Record.Attrs (inline)
         0     0% 26.07%      0.02s  0.37%  log/slog.appendJSONMarshal
         0     0% 26.07%      0.01s  0.19%  log/slog.appendJSONTime
         0     0% 26.07%      0.03s  0.56%  log/slog.appendJSONValue
         0     0% 26.07%     -0.02s  0.37%  log/slog/internal/buffer.(*Buffer).Free (inline)
         0     0% 26.07%     -0.01s  0.19%  log/slog/internal/buffer.init.func1
         0     0% 26.07%     -0.03s  0.56%  net.(*Dialer).DialContext
         0     0% 26.07%     -0.01s  0.19%  net.(*Resolver).internetAddrList
         0     0% 26.07%     -0.04s  0.74%  net.(*Resolver).lookupIP
         0     0% 26.07%     -0.01s  0.19%  net.(*Resolver).lookupIPAddr
         0     0% 26.07%     -0.04s  0.74%  net.(*Resolver).lookupIPAddr.func1
         0     0% 26.07%     -0.01s  0.19%  net.(*Resolver).resolveAddrList
         0     0% 26.07%      0.01s  0.19%  net.(*TCPAddr).sockaddr
         0     0% 26.07%      0.02s  0.37%  net.(*TCPConn).SetKeepAliveConfig
         0     0% 26.07%     -0.05s  0.93%  net.(*conf).hostLookupOrder
         0     0% 26.07%     -0.05s  0.93%  net.(*conf).lookupOrder
         0     0% 26.07%      0.04s  0.74%  net.(*conn).Close
         0     0% 26.07%      0.04s  0.74%  net.(*conn).Read
         0     0% 26.07%     -0.02s  0.37%  net.(*conn).SetReadDeadline
         0     0% 26.07%      0.01s  0.19%  net.(*conn).Write
         0     0% 26.07%      0.04s  0.74%  net.(*netFD).Close
         0     0% 26.07%     -0.02s  0.37%  net.(*netFD).SetReadDeadline (inline)
         0     0% 26.07%      0.01s  0.19%  net.(*netFD).Write
         0     0% 26.07%     -0.05s  0.93%  net.(*netFD).connect
         0     0% 26.07%     -0.04s  0.74%  net.(*netFD).dial
         0     0% 26.07%     -0.02s  0.37%  net.(*resolverConfig).tryUpdate
         0     0% 26.07%     -0.01s  0.19%  net.(*sysDialer).dialParallel
         0     0% 26.07%     -0.01s  0.19%  net.(*sysDialer).dialSerial
         0     0% 26.07%     -0.01s  0.19%  net.(*sysDialer).dialSingle
         0     0% 26.07%     -0.01s  0.19%  net.(*sysDialer).dialTCP
         0     0% 26.07%     -0.01s  0.19%  net.(*sysDialer).doDialTCP (inline)
         0     0% 26.07%     -0.01s  0.19%  net.(*sysDialer).doDialTCPProto
         0     0% 26.07%      0.04s  0.74%  net._C2func_getaddrinfo
         0     0% 26.07%      0.03s  0.56%  net._C_getaddrinfo
         0     0% 26.07%      0.03s  0.56%  net._C_getaddrinfo.func1 (inline)
         0     0% 26.07%      0.01s  0.19%  net.acquireThread
         0     0% 26.07%     -0.01s  0.19%  net.addrList.partition (inline)
         0     0% 26.07%      0.03s  0.56%  net.cgoLookupHostIP
         0     0% 26.07%      0.01s  0.19%  net.cgoLookupIP
         0     0% 26.07%      0.03s  0.56%  net.cgoLookupIP.func1
         0     0% 26.07%      0.01s  0.19%  net.doBlockingWithCtx[go.shape.[]net.IPAddr]
         0     0% 26.07%      0.05s  0.93%  net.doBlockingWithCtx[go.shape.[]net.IPAddr].func1
         0     0% 26.07%     -0.02s  0.37%  net.getSystemDNSConfig
         0     0% 26.07%     -0.04s  0.74%  net.init.func1
         0     0% 26.07%     -0.05s  0.93%  net.internetSocket
         0     0% 26.07%      0.01s  0.19%  net.ipToSockaddr
         0     0% 26.07%     -0.01s  0.19%  net.isIPv4
         0     0% 26.07%      0.04s  0.74%  net.newTCPConn
         0     0% 26.07%      0.01s  0.19%  net.setKeepAliveIdle
         0     0% 26.07%      0.01s  0.19%  net.setKeepAliveInterval
         0     0% 26.07%      0.01s  0.19%  net.setNoDelay
         0     0% 26.07%     -0.05s  0.93%  net.socket
         0     0% 26.07%     -0.01s  0.19%  net.sysSocket
         0     0% 26.07%     -0.02s  0.37%  net/http.(*Client).Do (inline)
         0     0% 26.07%     -0.02s  0.37%  net/http.(*Client).do
         0     0% 26.07%     -0.02s  0.37%  net/http.(*Client).send
         0     0% 26.07%     -0.01s  0.19%  net/http.(*Request).WithContext (inline)
         0     0% 26.07%     -0.07s  1.30%  net/http.(*Request).write
         0     0% 26.07%     -0.02s  0.37%  net/http.(*Transport).RoundTrip
         0     0% 26.07%     -0.11s  2.05%  net/http.(*Transport).dialConn
         0     0% 26.07%     -0.11s  2.05%  net/http.(*Transport).dialConnFor
         0     0% 26.07%     -0.01s  0.19%  net/http.(*Transport).getConn
         0     0% 26.07%     -0.02s  0.37%  net/http.(*Transport).queueForDial
         0     0% 26.07%     -0.02s  0.37%  net/http.(*Transport).roundTrip
         0     0% 26.07%     -0.02s  0.37%  net/http.(*Transport).startDialConnForLocked
         0     0% 26.07%     -0.11s  2.05%  net/http.(*Transport).startDialConnForLocked.func1
         0     0% 26.07%     -0.02s  0.37%  net/http.(*bodyEOFSignal).Read
         0     0% 26.07%     -0.01s  0.19%  net/http.(*chunkWriter).Write
         0     0% 26.07%     -0.01s  0.19%  net/http.(*chunkWriter).writeHeader
         0     0% 26.07%     -0.03s  0.56%  net/http.(*conn).readRequest
         0     0% 26.07%     -0.30s  5.59%  net/http.(*conn).serve
         0     0% 26.07%      0.01s  0.19%  net/http.(*connReader).Read
         0     0% 26.07%     -0.01s  0.19%  net/http.(*connReader).abortPendingRead
         0     0% 26.07%     -0.02s  0.37%  net/http.(*connReader).startBackgroundRead
         0     0% 26.07%      0.01s  0.19%  net/http.(*persistConn).Read
         0     0% 26.07%      0.04s  0.74%  net/http.(*persistConn).close
         0     0% 26.07%      0.04s  0.74%  net/http.(*persistConn).closeLocked
         0     0% 26.07%      0.07s  1.30%  net/http.(*persistConn).readLoop
         0     0% 26.07%      0.04s  0.74%  net/http.(*persistConn).readLoop.func1
         0     0% 26.07%      0.01s  0.19%  net/http.(*persistConn).readResponse
         0     0% 26.07%     -0.02s  0.37%  net/http.(*persistConn).roundTrip
         0     0% 26.07%     -0.07s  1.30%  net/http.(*persistConn).writeLoop
         0     0% 26.07%     -0.06s  1.12%  net/http.(*response).finishRequest
         0     0% 26.07%     -0.03s  0.56%  net/http.(*transferWriter).doBodyCopy
         0     0% 26.07%     -0.03s  0.56%  net/http.(*transferWriter).writeBody
         0     0% 26.07%     -0.21s  3.91%  net/http.HandlerFunc.ServeHTTP
         0     0% 26.07%     -0.02s  0.37%  net/http.Header.WriteSubset (inline)
         0     0% 26.07%     -0.02s  0.37%  net/http.Header.writeSubset
         0     0% 26.07%      0.01s  0.19%  net/http.NewRequest (inline)
         0     0% 26.07%      0.01s  0.19%  net/http.NewRequestWithContext
         0     0% 26.07%      0.01s  0.19%  net/http.ReadResponse
         0     0% 26.07%     -0.04s  0.74%  net/http.checkConnErrorWriter.Write
         0     0% 26.07%     -0.02s  0.37%  net/http.getCopyBuf (inline)
         0     0% 26.07%     -0.02s  0.37%  net/http.init.func15
         0     0% 26.07%      0.01s  0.19%  net/http.newBufioReader
         0     0% 26.07%     -0.01s  0.19%  net/http.newTransferWriter
         0     0% 26.07%      0.03s  0.56%  net/http.persistConnWriter.Write
         0     0% 26.07%     -0.03s  0.56%  net/http.readRequest
         0     0% 26.07%     -0.01s  0.19%  net/http.readTransfer
         0     0% 26.07%     -0.02s  0.37%  net/http.send
         0     0% 26.07%     -0.22s  4.10%  net/http.serverHandler.ServeHTTP
         0     0% 26.07%     -0.01s  0.19%  net/netip.ParseAddr
         0     0% 26.07%      0.01s  0.19%  net/textproto.(*Reader).ReadLine (inline)
         0     0% 26.07%     -0.05s  0.93%  net/textproto.(*Reader).ReadMIMEHeader (inline)
         0     0% 26.07%     -0.01s  0.19%  net/textproto.(*Reader).readContinuedLineSlice
         0     0% 26.07%      0.01s  0.19%  net/textproto.(*Reader).readLineSlice
         0     0% 26.07%     -0.01s  0.19%  net/textproto.mustHaveFieldNameColon
         0     0% 26.07%     -0.05s  0.93%  net/textproto.readMIMEHeader
         0     0% 26.07%      0.01s  0.19%  net/url.Parse
         0     0% 26.07%      0.01s  0.19%  net/url.parse
         0     0% 26.07%      0.01s  0.19%  net/url.parseAuthority
         0     0% 26.07%      0.01s  0.19%  net/url.parseHost
         0     0% 26.07%      0.01s  0.19%  net/url.shouldEscape
         0     0% 26.07%      0.01s  0.19%  net/url.unescape
         0     0% 26.07%      0.03s  0.56%  os.(*File).Sync
         0     0% 26.07%      0.02s  0.37%  os.(*File).Write
         0     0% 26.07%      0.02s  0.37%  os.(*File).write (inline)
         0     0% 26.07%     -0.01s  0.19%  os.Stat
         0     0% 26.07%     -0.01s  0.19%  os.statNolog
         0     0% 26.07%     -0.01s  0.19%  reflect.(*rtype).Name
         0     0% 26.07%     -0.01s  0.19%  reflect.(*rtype).NumMethod
         0     0% 26.07%     -0.01s  0.19%  reflect.(*rtype).String
         0     0% 26.07%     -0.01s  0.19%  reflect.Value.Grow
         0     0% 26.07%     -0.01s  0.19%  reflect.Value.grow
         0     0% 26.07%     -0.01s  0.19%  reflect.growslice
         0     0% 26.07%      0.01s  0.19%  runtime.(*activeSweep).isDone (inline)
         0     0% 26.07%      0.01s  0.19%  runtime.(*atomicHeadTailIndex).cas (inline)
         0     0% 26.07%     -0.03s  0.56%  runtime.(*atomicHeadTailIndex).load (inline)
         0     0% 26.07%     -0.01s  0.19%  runtime.(*atomicScavChunkData).load
         0     0% 26.07%      0.01s  0.19%  runtime.(*atomicSpanSetSpinePointer).Load (inline)
         0     0% 26.07%     -0.01s  0.19%  runtime.(*gcCPULimiterState).update
         0     0% 26.07%     -0.01s  0.19%  runtime.(*gcCPULimiterState).updateLocked
         0     0% 26.07%     -0.02s  0.37%  runtime.(*gcControllerState).enlistWorker
         0     0% 26.07%     -0.02s  0.37%  runtime.(*gcControllerState).heapGoal (inline)
         0     0% 26.07%     -0.02s  0.37%  runtime.(*gcControllerState).heapGoalInternal
         0     0% 26.07%      0.01s  0.19%  runtime.(*gcControllerState).needIdleMarkWorker
         0     0% 26.07%     -0.02s  0.37%  runtime.(*gcControllerState).revise
         0     0% 26.07%     -0.02s  0.37%  runtime.(*gcControllerState).update
         0     0% 26.07%     -0.02s  0.37%  runtime.(*gcWork).balance
         0     0% 26.07%      0.01s  0.19%  runtime.(*gcWork).init
         0     0% 26.07%     -0.01s  0.19%  runtime.(*inlineUnwinder).resolveInternal (inline)
         0     0% 26.07%      0.01s  0.19%  runtime.(*mSpanStateBox).get (inline)
         0     0% 26.07%     -0.06s  1.12%  runtime.(*mcache).prepareForSweep
         0     0% 26.07%     -0.01s  0.19%  runtime.(*mcache).refill
         0     0% 26.07%     -0.05s  0.93%  runtime.(*mcache).releaseAll
         0     0% 26.07%      0.01s  0.19%  runtime.(*mcentral).cacheSpan
         0     0% 26.07%     -0.01s  0.19%  runtime.(*mheap).alloc
         0     0% 26.07%     -0.01s  0.19%  runtime.(*mheap).alloc.func1
         0     0% 26.07%     -0.01s  0.19%  runtime.(*mheap).allocMSpanLocked
         0     0% 26.07%     -0.04s  0.74%  runtime.(*mheap).allocSpan
         0     0% 26.07%      0.02s  0.37%  runtime.(*mheap).freeSpan (inline)
         0     0% 26.07%      0.03s  0.56%  runtime.(*mheap).freeSpanLocked
         0     0% 26.07%     -0.02s  0.37%  runtime.(*mheap).initSpan
         0     0% 26.07%     -0.03s  0.56%  runtime.(*mheap).nextSpanForSweep
         0     0% 26.07%      0.03s  0.56%  runtime.(*mheap).reclaim
         0     0% 26.07%      0.03s  0.56%  runtime.(*mheap).reclaimChunk
         0     0% 26.07%     -0.01s  0.19%  runtime.(*mspan).countAlloc
         0     0% 26.07%     -0.01s  0.19%  runtime.(*mspan).ensureSwept
         0     0% 26.07%     -0.05s  0.93%  runtime.(*mspan).markBitsForIndex (inline)
         0     0% 26.07%     -0.01s  0.19%  runtime.(*pageAlloc).alloc
         0     0% 26.07%     -0.02s  0.37%  runtime.(*pageAlloc).allocRange
         0     0% 26.07%      0.01s  0.19%  runtime.(*pageAlloc).find
         0     0% 26.07%     -0.02s  0.37%  runtime.(*pageAlloc).scavenge.func1
         0     0% 26.07%     -0.02s  0.37%  runtime.(*pageAlloc).scavengeOne
         0     0% 26.07%      0.01s  0.19%  runtime.(*pallocBits).summarize
         0     0% 26.07%     -0.01s  0.19%  runtime.(*scavengeIndex).alloc
         0     0% 26.07%     -0.03s  0.56%  runtime.(*scavengerState).init.func2
         0     0% 26.07%     -0.03s  0.56%  runtime.(*scavengerState).run
         0     0% 26.07%     -0.02s  0.37%  runtime.(*spanSet).pop
         0     0% 26.07%     -0.01s  0.19%  runtime.(*spanSetBlockAlloc).free
         0     0% 26.07%      0.01s  0.19%  runtime.(*stackScanState).addObject
         0     0% 26.07%      0.02s  0.37%  runtime.(*sweepLocked).sweep.(*mheap).freeSpan.func3
         0     0% 26.07%      0.01s  0.19%  runtime.(*sysMemStat).load (inline)
         0     0% 26.07%      0.01s  0.19%  runtime.(*timeHistogram).record
         0     0% 26.07%     -0.02s  0.37%  runtime.(*timers).check
         0     0% 26.07%      0.01s  0.19%  runtime.(*unwinder).initAt
         0     0% 26.07%     -0.04s  0.74%  runtime.(*unwinder).next
         0     0% 26.07%      0.01s  0.19%  runtime.(*unwinder).symPC
         0     0% 26.07%     -0.02s  0.37%  runtime.Callers (inline)
         0     0% 26.07%     -0.05s  0.93%  runtime.acquirep
         0     0% 26.07%     -0.02s  0.37%  runtime.addspecial
         0     0% 26.07%     -0.02s  0.37%  runtime.adjustframe
         0     0% 26.07%     -0.03s  0.56%  runtime.bgscavenge
         0     0% 26.07%      0.03s  0.56%  runtime.callers
         0     0% 26.07%      0.03s  0.56%  runtime.callers.func1
         0     0% 26.07%      0.01s  0.19%  runtime.castogscanstatus
         0     0% 26.07%     -0.01s  0.19%  runtime.cgoCheckArg
         0     0% 26.07%     -0.01s  0.19%  runtime.cgoCheckPointer
         0     0% 26.07%     -0.01s  0.19%  runtime.chanrecv1
         0     0% 26.07%      0.01s  0.19%  runtime.chansend1
         0     0% 26.07%      0.01s  0.19%  runtime.checkIdleGCNoP
         0     0% 26.07%     -0.01s  0.19%  runtime.clearpools
         0     0% 26.07%     -0.03s  0.56%  runtime.convT
         0     0% 26.07%     -0.04s  0.74%  runtime.convTstring
         0     0% 26.07%      0.02s  0.37%  runtime.deductSweepCredit
         0     0% 26.07%     -0.01s  0.19%  runtime.dodiv
         0     0% 26.07%      0.01s  0.19%  runtime.entersyscall
         0     0% 26.07%     -0.01s  0.19%  runtime.execute
         0     0% 26.07%     -0.01s  0.19%  runtime.exitsyscall0
         0     0% 26.07%     -0.28s  5.21%  runtime.findRunnable
         0     0% 26.07%     -0.02s  0.37%  runtime.finishsweep_m
         0     0% 26.07%      0.01s  0.19%  runtime.float64toint64
         0     0% 26.07%     -0.01s  0.19%  runtime.forEachG
         0     0% 26.07%     -0.03s  0.56%  runtime.forEachP (inline)
         0     0% 26.07%     -0.03s  0.56%  runtime.forEachPInternal
         0     0% 26.07%     -0.01s  0.19%  runtime.freeSomeWbufs
         0     0% 26.07%     -0.01s  0.19%  runtime.freeSpecial
         0     0% 26.07%     -0.09s  1.68%  runtime.futexsleep
         0     0% 26.07%     -0.06s  1.12%  runtime.futexwakeup
         0     0% 26.07%     -0.22s  4.10%  runtime.gcAssistAlloc
         0     0% 26.07%     -0.24s  4.47%  runtime.gcAssistAlloc.func2
         0     0% 26.07%     -0.24s  4.47%  runtime.gcAssistAlloc1
         0     0% 26.07%     -0.39s  7.26%  runtime.gcBgMarkWorker
         0     0% 26.07%     -0.30s  5.59%  runtime.gcBgMarkWorker.func2
         0     0% 26.07%     -0.29s  5.40%  runtime.gcDrainMarkWorkerDedicated (inline)
         0     0% 26.07%     -0.02s  0.37%  runtime.gcMarkDone.forEachP.func5
         0     0% 26.07%     -0.01s  0.19%  runtime.gcMarkDone.func2
         0     0% 26.07%     -0.01s  0.19%  runtime.gcMarkDone.func3
         0     0% 26.07%     -0.05s  0.93%  runtime.gcMarkTermination
         0     0% 26.07%     -0.01s  0.19%  runtime.gcMarkTermination.forEachP.func6
         0     0% 26.07%     -0.02s  0.37%  runtime.gcMarkTermination.func3
         0     0% 26.07%     -0.01s  0.19%  runtime.gcResetMarkState
         0     0% 26.07%     -0.05s  0.93%  runtime.gcStart
         0     0% 26.07%      0.01s  0.19%  runtime.gcStart.func2
         0     0% 26.07%     -0.02s  0.37%  runtime.gcStart.func3
         0     0% 26.07%     -0.02s  0.37%  runtime.gcStart.func4
         0     0% 26.07%     -0.01s  0.19%  runtime.gcTrigger.test
         0     0% 26.07%     -0.04s  0.74%  runtime.gcstopm
         0     0% 26.07%     -0.02s  0.37%  runtime.getGCMaskOnDemand
         0     0% 26.07%      0.01s  0.19%  runtime.gfget
         0     0% 26.07%      0.01s  0.19%  runtime.gfget.func2
         0     0% 26.07%      0.01s  0.19%  runtime.globrunqget
         0     0% 26.07%     -0.11s  2.05%  runtime.goexit0
         0     0% 26.07%     -0.02s  0.37%  runtime.gopreempt_m (inline)
         0     0% 26.07%      0.01s  0.19%  runtime.goready (inline)
         0     0% 26.07%     -0.03s  0.56%  runtime.goschedImpl
         0     0% 26.07%     -0.01s  0.19%  runtime.goschedguarded_m
         0     0% 26.07%      0.01s  0.19%  runtime.handoffp
         0     0% 26.07%     -0.01s  0.19%  runtime.injectglist
         0     0% 26.07%     -0.01s  0.19%  runtime.injectglist.func1
         0     0% 26.07%      0.01s  0.19%  runtime.int64tofloat64
         0     0% 26.07%      0.01s  0.19%  runtime.isSweepDone (inline)
         0     0% 26.07%     -0.08s  1.49%  runtime.lock (partial-inline)
         0     0% 26.07%     -0.09s  1.68%  runtime.mPark (inline)
         0     0% 26.07%     -0.01s  0.19%  runtime.mProf_Flush
         0     0% 26.07%     -0.01s  0.19%  runtime.mProf_FlushLocked
         0     0% 26.07%      0.03s  0.56%  runtime.mProf_Malloc
         0     0% 26.07%     -0.02s  0.37%  runtime.mProf_Malloc.func1
         0     0% 26.07%     -0.02s  0.37%  runtime.makechan
         0     0% 26.07%     -0.01s  0.19%  runtime.makemap
         0     0% 26.07%     -0.04s  0.74%  runtime.makemap_small
         0     0% 26.07%     -0.30s  5.59%  runtime.mallocgc
         0     0% 26.07%     -0.03s  0.56%  runtime.mallocgcLarge
         0     0% 26.07%     -0.01s  0.19%  runtime.mapIterStart
         0     0% 26.07%      0.01s  0.19%  runtime.mapdelete_faststr
         0     0% 26.07%     -0.01s  0.19%  runtime.markBits.setMarked (inline)
         0     0% 26.07%     -0.02s  0.37%  runtime.markroot
         0     0% 26.07%      0.02s  0.37%  runtime.markrootBlock
         0     0% 26.07%      0.01s  0.19%  runtime.markrootSpans
         0     0% 26.07%     -0.30s  5.59%  runtime.mcall
         0     0% 26.07%     -0.02s  0.37%  runtime.memclrNoHeapPointersChunked
         0     0% 26.07%     -0.02s  0.37%  runtime.morestack
         0     0% 26.07%     -0.05s  0.93%  runtime.netpoll
         0     0% 26.07%      0.01s  0.19%  runtime.netpollclose
         0     0% 26.07%     -0.01s  0.19%  runtime.netpollopen
         0     0% 26.07%     -0.02s  0.37%  runtime.newArenaMayUnlock
         0     0% 26.07%     -0.01s  0.19%  runtime.newInlineUnwinder
         0     0% 26.07%     -0.05s  0.93%  runtime.newMarkBits
         0     0% 26.07%     -0.01s  0.19%  runtime.newarray
         0     0% 26.07%     -0.11s  2.05%  runtime.newobject
         0     0% 26.07%     -0.02s  0.37%  runtime.newproc
         0     0% 26.07%     -0.02s  0.37%  runtime.newproc.func1
         0     0% 26.07%     -0.07s  1.30%  runtime.newstack
         0     0% 26.07%     -0.09s  1.68%  runtime.notesleep
         0     0% 26.07%     -0.05s  0.93%  runtime.notewakeup
         0     0% 26.07%     -0.01s  0.19%  runtime.pageIndexOf (inline)
         0     0% 26.07%     -0.17s  3.17%  runtime.park_m
         0     0% 26.07%     -0.01s  0.19%  runtime.pcdatavalue
         0     0% 26.07%     -0.01s  0.19%  runtime.pcdatavalue1
         0     0% 26.07%      0.01s  0.19%  runtime.pidlegetSpinning
         0     0% 26.07%     -0.01s  0.19%  runtime.preemptM
         0     0% 26.07%     -0.01s  0.19%  runtime.preemptone
         0     0% 26.07%     -0.01s  0.19%  runtime.procresize
         0     0% 26.07%      0.03s  0.56%  runtime.profilealloc
         0     0% 26.07%     -0.01s  0.19%  runtime.rand
         0     0% 26.07%      0.01s  0.19%  runtime.reentersyscall
         0     0% 26.07%     -0.02s  0.37%  runtime.runSafePointFn
         0     0% 26.07%     -0.02s  0.37%  runtime.runqsteal
         0     0% 26.07%     -0.07s  1.30%  runtime.scanstack
         0     0% 26.07%     -0.29s  5.40%  runtime.schedule
         0     0% 26.07%      0.02s  0.37%  runtime.selectgo
         0     0% 26.07%     -0.01s  0.19%  runtime.selectnbsend
         0     0% 26.07%     -0.01s  0.19%  runtime.semawakeup
         0     0% 26.07%      0.01s  0.19%  runtime.send.goready.func1
         0     0% 26.07%     -0.02s  0.37%  runtime.setprofilebucket
         0     0% 26.07%     -0.01s  0.19%  runtime.signalM
         0     0% 26.07%     -0.01s  0.19%  runtime.slicebytetostring
         0     0% 26.07%     -0.01s  0.19%  runtime.spanOf (inline)
         0     0% 26.07%      0.01s  0.19%  runtime.stackalloc
         0     0% 26.07%     -0.01s  0.19%  runtime.stackcache_clear
         0     0% 26.07%      0.01s  0.19%  runtime.stackcacherefill
         0     0% 26.07%      0.01s  0.19%  runtime.stackpoolalloc
         0     0% 26.07%     -0.04s  0.74%  runtime.startTheWorldWithSema
         0     0% 26.07%     -0.04s  0.74%  runtime.startm
         0     0% 26.07%     -0.15s  2.79%  runtime.stopm
         0     0% 26.07%     -0.05s  0.93%  runtime.sweepone
         0     0% 26.07%      0.01s  0.19%  runtime.sysUnused
         0     0% 26.07%      0.01s  0.19%  runtime.sysUnusedOS
         0     0% 26.07%     -0.65s 12.10%  runtime.systemstack
         0     0% 26.07%     -0.01s  0.19%  runtime.typedmemmove
         0     0% 26.07%      0.01s  0.19%  runtime.uint64tofloat64 (inline)
         0     0% 26.07%     -0.01s  0.19%  runtime.unlock (inline)
         0     0% 26.07%     -0.01s  0.19%  runtime.unlock2Wake
         0     0% 26.07%     -0.01s  0.19%  runtime.unlockWithRank (inline)
         0     0% 26.07%     -0.03s  0.56%  runtime.wakep
         0     0% 26.07%     -0.01s  0.19%  runtime.wbBufFlush
         0     0% 26.07%     -0.01s  0.19%  runtime.wbBufFlush.func1
         0     0% 26.07%     -0.01s  0.19%  runtime.wbBufFlush1
         0     0% 26.07%      0.01s  0.19%  runtime.wbMove
         0     0% 26.07%      0.01s  0.19%  slices.SortFunc[go.shape.[]encoding/json.reflectWithString,go.shape.struct { encoding/json.v reflect.Value; encoding/json.ks string }]
         0     0% 26.07%      0.01s  0.19%  slices.insertionSortCmpFunc[go.shape.struct { encoding/json.v reflect.Value; encoding/json.ks string }] (inline)
         0     0% 26.07%      0.01s  0.19%  slices.pdqsortCmpFunc[go.shape.struct { encoding/json.v reflect.Value; encoding/json.ks string }]
         0     0% 26.07%      0.01s  0.19%  strconv.ParseFloat
         0     0% 26.07%      0.01s  0.19%  strconv.atof64
         0     0% 26.07%      0.01s  0.19%  strconv.parseFloatPrefix
         0     0% 26.07%     -0.02s  0.37%  strings.(*Builder).Grow
         0     0% 26.07%     -0.01s  0.19%  strings.(*Builder).WriteString (inline)
         0     0% 26.07%     -0.02s  0.37%  strings.(*Builder).grow
         0     0% 26.07%     -0.03s  0.56%  strings.Join
         0     0% 26.07%     -0.01s  0.19%  strings.Split (inline)
         0     0% 26.07%     -0.01s  0.19%  strings.SplitN (inline)
         0     0% 26.07%     -0.02s  0.37%  strings.genSplit
         0     0% 26.07%     -0.01s  0.19%  sync.(*Map).Load (inline)
         0     0% 26.07%     -0.03s  0.56%  sync.(*Once).Do
         0     0% 26.07%     -0.02s  0.37%  sync.(*Pool).Get
         0     0% 26.07%     -0.02s  0.37%  sync.(*Pool).Put
         0     0% 26.07%     -0.01s  0.19%  sync.(*Pool).getSlow
         0     0% 26.07%      0.01s  0.19%  sync.(*Pool).pinSlow
         0     0% 26.07%     -0.01s  0.19%  sync.(*poolChain).popTail
         0     0% 26.07%     -0.02s  0.37%  sync.(*poolChain).pushHead
         0     0% 26.07%     -0.01s  0.19%  sync.(*poolDequeue).pushHead
         0     0% 26.07%     -0.01s  0.19%  sync/atomic.(*Uint64).Add (inline)
         0     0% 26.07%      0.01s  0.19%  sync/atomic.(*Value).Store
         0     0% 26.07%      0.01s  0.19%  sync/atomic.CompareAndSwapPointer
         0     0% 26.07%      0.04s  0.74%  syscall.Close
         0     0% 26.07%     -0.05s  0.93%  syscall.Connect
         0     0% 26.07%      0.03s  0.56%  syscall.Fsync
         0     0% 26.07%      0.01s  0.19%  syscall.Getpeername
         0     0% 26.07%      0.09s  1.68%  syscall.RawSyscall6
         0     0% 26.07%      0.05s  0.93%  syscall.Read (inline)
         0     0% 26.07%      0.03s  0.56%  syscall.SetsockoptInt (inline)
         0     0% 26.07%     -0.01s  0.19%  syscall.Socket
         0     0% 26.07%      0.10s  1.86%  syscall.Syscall
         0     0% 26.07%     -0.05s  0.93%  syscall.connect
         0     0% 26.07%      0.01s  0.19%  syscall.getpeername
         0     0% 26.07%      0.05s  0.93%  syscall.read
         0     0% 26.07%      0.03s  0.56%  syscall.setsockopt
         0     0% 26.07%     -0.01s  0.19%  syscall.socket
         0     0% 26.07%     -0.01s  0.19%  time.Time.Add
         0     0% 26.07%      0.01s  0.19%  time.Time.AppendFormat
         0     0% 26.07%     -0.01s  0.19%  time.Time.Sub
         0     0% 26.07%      0.01s  0.19%  time.Time.appendFormatRFC3339
         0     0% 26.07%     -0.02s  0.37%  time.Until
         0     0% 26.07%      0.01s  0.19%  time.absDays.date
         0     0% 26.07%      0.01s  0.19%  time.absDays.split (inline)
         0     0% 26.07%     -0.01s  0.19%  time.runtimeNano
