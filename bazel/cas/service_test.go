package cas

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/golang/protobuf/proto"
	uuid "github.com/nu7hatch/gouuid"
	remoteexecution "github.com/twitter/scoot/bazel/remoteexecution"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/bytestream"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/twitter/scoot/bazel"
	"github.com/twitter/scoot/common/stats"
	"github.com/twitter/scoot/snapshot/store"
)

var testHash1 string = "ce58a4479be1d32816ee82e57eae04415dc2bda173fa7b0f11d18aa67856f242"
var testSize1 int64 = 7
var testData1 []byte = []byte("abc1234")

func TestFindMissingBlobs(t *testing.T) {
	f := &store.FakeStore{}
	s := casServer{storeConfig: &store.StoreConfig{Store: f}, stat: stats.NilStatsReceiver()}

	// Create 2 digests, write 1 to Store, check both for missing, expect other 1 back
	dExists := &remoteexecution.Digest{Hash: "abc123", SizeBytes: 1}
	dMissing := &remoteexecution.Digest{Hash: "efg456", SizeBytes: 9}
	digests := []*remoteexecution.Digest{dExists, dMissing}
	expected := []*remoteexecution.Digest{dMissing}

	resourceName := bazel.DigestStoreName(dExists)
	err := f.Write(resourceName, bytes.NewReader([]byte("")), nil)
	if err != nil {
		t.Fatalf("Failed to write into FakeStore: %v", err)
	}

	req := &remoteexecution.FindMissingBlobsRequest{BlobDigests: digests}

	res, err := s.FindMissingBlobs(context.Background(), req)
	if err != nil {
		t.Fatalf("Error response from FindMissingBlobs: %v", err)
	}

	if len(expected) != len(res.MissingBlobDigests) {
		t.Fatalf("Length of missing blobs mismatch, expected %d got %d", len(expected), len(res.MissingBlobDigests))
	}
	for i, d := range res.MissingBlobDigests {
		if expected[i] != d {
			t.Fatalf("Non-match iterating through missing digests, expected %s got: %s", expected[i], d)
		}
	}
}

func TestRead(t *testing.T) {
	f := &store.FakeStore{}
	s := casServer{storeConfig: &store.StoreConfig{Store: f}, stat: stats.NilStatsReceiver()}

	// Write a resource to underlying store
	d := &remoteexecution.Digest{Hash: testHash1, SizeBytes: testSize1}
	resourceName := bazel.DigestStoreName(d)
	err := f.Write(resourceName, bytes.NewReader(testData1), nil)
	if err != nil {
		t.Fatalf("Failed to write into FakeStore: %v", err)
	}

	// Make a ReadRequest that exercises read limits and offsets
	offset, limit := int64(2), int64(2)
	req := &bytestream.ReadRequest{ResourceName: fmt.Sprintf("blobs/%s/%d", testHash1, testSize1), ReadOffset: offset, ReadLimit: limit}
	r := makeFakeReadServer()

	// Make actual Read request
	err = s.Read(req, r)
	if err != nil {
		t.Fatalf("Error response from Read: %v", err)
	}

	// Get data sent/captured by fake server, and compare with expected based on testdata, limit, and offset
	b, err := ioutil.ReadAll(r.buffer)
	if err != nil {
		t.Fatalf("Error reading from fake server data: %v", err)
	}
	if bytes.Compare(b, testData1[offset:]) != 0 {
		t.Fatalf("Data read from fake server did not match - expected: %s, got: %s", testData1[offset:], b)
	}
	sends := int((testSize1-offset)/limit + (testSize1-offset)%limit)
	if r.sendCount != sends {
		t.Fatalf("Fake server Send() count mismatch - expected %d times based on - data len: %d ReadOffset: %d ReadLimit %d. got: %d", sends, testSize1, offset, limit, r.sendCount)
	}
	r.reset()
}

func TestReadEmpty(t *testing.T) {
	f := &store.FakeStore{}
	s := casServer{storeConfig: &store.StoreConfig{Store: f}, stat: stats.NilStatsReceiver()}

	// Note: don't actually write the underlying resource beforehand. We expect
	// that reading an empty blob will bypass the underlying store

	req := &bytestream.ReadRequest{ResourceName: fmt.Sprintf("blobs/%s/-1", bazel.EmptySha), ReadOffset: int64(0), ReadLimit: int64(0)}
	r := makeFakeReadServer()

	// Make actual Read request
	err := s.Read(req, r)
	if err != nil {
		t.Fatalf("Error response from Read: %v", err)
	}

	// Get data sent/captured by fake server, and compare with expected based on testdata, limit, and offset
	b, err := ioutil.ReadAll(r.buffer)
	if err != nil {
		t.Fatalf("Error reading from fake server data: %v", err)
	}
	if bytes.Compare(b, []byte{}) != 0 {
		t.Fatalf("Data read from fake server did not match - expected: %s, got: %s", []byte{}, b)
	}
	r.reset()
}

func TestReadMissing(t *testing.T) {
	f := &store.FakeStore{}
	s := casServer{storeConfig: &store.StoreConfig{Store: f}, stat: stats.NilStatsReceiver()}

	// Note: don't actually write the underlying resource beforehand. We expect
	// that reading an empty blob will bypass the underlying store

	req := &bytestream.ReadRequest{ResourceName: fmt.Sprintf("blobs/%s/%d", testHash1, testSize1)}
	r := makeFakeReadServer()

	// Make actual Read request
	err := s.Read(req, r)
	if err == nil {
		t.Fatal("Unexpected success - want NotFound error response from Read")
	}
	c := status.Code(err)
	if c != codes.NotFound {
		t.Fatalf("Status code from error not expected: %s, got: %s", codes.NotFound, c)
	}

	r.reset()
}

func TestWrite(t *testing.T) {
	f := &store.FakeStore{}
	s := casServer{storeConfig: &store.StoreConfig{Store: f}, stat: stats.NilStatsReceiver()}

	w := makeFakeWriteServer(testHash1, testSize1, testData1, 3)

	// Make Write request with test data
	err := s.Write(w)
	if err != nil {
		t.Fatalf("Error response from Write: %v", err)
	}

	// Verify that fake write server was invoked as expected
	if w.committedSize != testSize1 {
		t.Fatalf("Size committed to fake server did not match - expected: %d, got: %d", testSize1, w.committedSize)
	}
	if w.recvCount != w.recvChunks {
		t.Fatalf("Number of write chunks to fake server did not match - expected: %d, got: %d", w.recvChunks, w.recvCount)
	}

	// Verify Write by reading directly from underlying Store
	d := &remoteexecution.Digest{Hash: testHash1, SizeBytes: testSize1}
	resourceName := bazel.DigestStoreName(d)
	r, err := f.OpenForRead(resourceName)
	if err != nil {
		t.Fatalf("Failed to open expected resource for reading: %s: %v", resourceName, err)
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatalf("Error reading from fake store: %v", err)
	}
	if bytes.Compare(b, testData1) != 0 {
		t.Fatalf("Data read from store did not match - expected: %s, got: %s", testData1, b)
	}
}

func TestWriteEmpty(t *testing.T) {
	f := &store.FakeStore{}
	s := casServer{storeConfig: &store.StoreConfig{Store: f}, stat: stats.NilStatsReceiver()}

	w := makeFakeWriteServer(bazel.EmptySha, bazel.EmptySize, []byte{}, 1)

	// Make Write request with test data
	err := s.Write(w)
	if err != nil {
		t.Fatalf("Error response from Write: %v", err)
	}

	// Verify that fake write server was invoked as expected
	if w.committedSize != bazel.EmptySize {
		t.Fatalf("Size committed to fake server did not match - expected: %d, got: %d", bazel.EmptySize, w.committedSize)
	}
	if w.recvCount != w.recvChunks {
		t.Fatalf("Number of write chunks to fake server did not match - expected: %d, got: %d", w.recvChunks, w.recvCount)
	}
	// Don't verify - we reserve the right to not actually write to the underlying store in this scenario
}

func TestWriteExisting(t *testing.T) {
	f := &store.FakeStore{}
	s := casServer{storeConfig: &store.StoreConfig{Store: f}, stat: stats.NilStatsReceiver()}

	// Pre-write data directly to underlying Store
	d := &remoteexecution.Digest{Hash: testHash1, SizeBytes: testSize1}

	resourceName := bazel.DigestStoreName(d)
	err := f.Write(resourceName, bytes.NewReader(testData1), nil)
	if err != nil {
		t.Fatalf("Failed to write into FakeStore: %v", err)
	}

	w := makeFakeWriteServer(testHash1, testSize1, testData1, 3)

	// Make Write request with test data matching pre-written data
	err = s.Write(w)
	if err != nil {
		t.Fatalf("Error response from Write: %v", err)
	}

	// Verify that fake write server was invoked as expected - recv was called only once
	if w.committedSize != testSize1 {
		t.Fatalf("Size committed to fake server did not match - expected: %d, got: %d", testSize1, w.committedSize)
	}
	if w.recvCount != 1 {
		t.Fatalf("Number of write chunks to fake server did not match - expected: %d, got: %d", 1, w.recvCount)
	}
}

func TestQueryWriteStatusStub(t *testing.T) {
	f := &store.FakeStore{}
	s := casServer{storeConfig: &store.StoreConfig{Store: f}, stat: stats.NilStatsReceiver()}

	req := &bytestream.QueryWriteStatusRequest{}

	_, err := s.QueryWriteStatus(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error response from QueryWriteStatus, got nil")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("Not ok reading grpc status from error")
	}
	if st.Code() != codes.Unimplemented {
		t.Fatalf("Expected status code %d, got: %d", codes.Unimplemented, st.Code())
	}
}

func TestBatchUpdateBlobsStub(t *testing.T) {
	s := casServer{stat: stats.NilStatsReceiver()}
	req := &remoteexecution.BatchUpdateBlobsRequest{}

	_, err := s.BatchUpdateBlobs(context.Background(), req)
	if err == nil {
		t.Fatalf("Non-error response from BatchUpdateBlobs")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("Not ok reading grpc status from error")
	}
	if st.Code() != codes.Unimplemented {
		t.Fatalf("Expected status code %d, got: %d", codes.Unimplemented, st.Code())
	}
}

func TestGetTreeStub(t *testing.T) {
	s := casServer{stat: stats.NilStatsReceiver()}
	req := &remoteexecution.GetTreeRequest{}
	gtServer := &fakeGetTreeServer{}

	err := s.GetTree(req, gtServer)
	if err == nil {
		t.Fatalf("Non-error response from GetTree")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("Not ok reading grpc status from error")
	}
	if st.Code() != codes.Unimplemented {
		t.Fatalf("Expected status code %d, got: %d", codes.Unimplemented, st.Code())
	}
}

func TestMakeResultAddress(t *testing.T) {
	ad := &remoteexecution.Digest{Hash: testHash1, SizeBytes: testSize1}
	// e.g. `echo -n "<testHash1>-<ResultAddressKey>" | shasum -a 256`
	knownHash := "fdc8c407bc2aa6d6cb514ace4299b2f414c4476a77123e3557dafd103d889124"
	storeName := fmt.Sprintf("%s-%s.%s", bazel.StorePrefix, knownHash, bazel.StorePrefix)

	resultAddr, err := makeCacheResultAddress(ad)
	if err != nil {
		t.Fatalf("Failed to create cache result address: %v", err)
	}
	if resultAddr.storeName != storeName {
		t.Fatalf("Unexpected resulting store name: %s, want: %s", resultAddr.storeName, storeName)
	}
}

func TestGetActionResult(t *testing.T) {
	f := &store.FakeStore{}
	s := casServer{storeConfig: &store.StoreConfig{Store: f}, stat: stats.NilStatsReceiver()}

	arAsBytes, err := getFakeActionResult()
	if err != nil {
		t.Fatalf("Error getting ActionResult: %s", err)
	}

	// Get ActionDigest and Write AR to underlying store using our result cache addressing convention
	ad := &remoteexecution.Digest{Hash: testHash1, SizeBytes: testSize1}
	address, err := makeCacheResultAddress(ad)
	if err != nil {
		t.Fatalf("Failed to create cache result adress: %v", err)
	}

	err = f.Write(address.storeName, bytes.NewReader(arAsBytes), nil)
	if err != nil {
		t.Fatalf("Failed to write into FakeStore: %v", err)
	}

	// Make GetActionResult request
	req := &remoteexecution.GetActionResultRequest{ActionDigest: ad}

	resAr, err := s.GetActionResult(context.Background(), req)
	if err != nil {
		t.Fatalf("Error from GetActionResult: %v", err)
	}

	// Convert result to bytes and compare
	resAsBytes, err := proto.Marshal(resAr)
	if err != nil {
		t.Fatalf("Error serializing result: %s", err)
	}

	if bytes.Compare(arAsBytes, resAsBytes) != 0 {
		t.Fatal("Result not as expected after serialization")
	}
}

func TestGetActionResultMissing(t *testing.T) {
	f := &store.FakeStore{}
	s := casServer{storeConfig: &store.StoreConfig{Store: f}, stat: stats.NilStatsReceiver()}

	// Make GetActionResult request
	req := &remoteexecution.GetActionResultRequest{
		ActionDigest: &remoteexecution.Digest{
			Hash:      testHash1,
			SizeBytes: testSize1,
		},
	}

	res, err := s.GetActionResult(context.Background(), req)
	if err == nil {
		t.Fatal("Unexpected success - want NotFound error response from GetActionResult")
	}
	if res != nil {
		t.Fatal("Unexpected non-nil GetActionResult")
	}
	c := status.Code(err)
	if c != codes.NotFound {
		t.Fatalf("Status code from error not expected: %s, got: %s", codes.NotFound, c)
	}
}

func TestUpdateActionResult(t *testing.T) {
	f := &store.FakeStore{}
	s := casServer{storeConfig: &store.StoreConfig{Store: f}, stat: stats.NilStatsReceiver()}

	rc := int32(42)
	ad := &remoteexecution.Digest{Hash: testHash1, SizeBytes: testSize1}
	ar := &remoteexecution.ActionResult{ExitCode: rc}
	req := &remoteexecution.UpdateActionResultRequest{ActionDigest: ad, ActionResult: ar}

	_, err := s.UpdateActionResult(context.Background(), req)
	if err != nil {
		t.Fatalf("Error from UpdateActionResult: %v", err)
	}

	// Read from underlying store
	address, err := makeCacheResultAddress(ad)
	if err != nil {
		t.Fatalf("Failed to create cache result adress: %v", err)
	}

	r, err := f.OpenForRead(address.storeName)
	if err != nil {
		t.Fatalf("Failed to open expected resource for reading: %s: %v", address.storeName, err)
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatalf("Error reading from fake store: %v", err)
	}
	resAr := &remoteexecution.ActionResult{}
	if err = proto.Unmarshal(b, resAr); err != nil {
		t.Fatalf("Error deserializing result: %s", err)
	}

	if resAr.GetExitCode() != rc {
		t.Fatalf("Result not as expected, got: %d, want: %d", resAr.GetExitCode(), rc)
	}
}

// Fake Read/Write Servers

// Read server that fakes sending data to a client.
// 	sendCount - tracks times the server invokes Send for verification
// Implements bytestream.ByteStream_ReadServer interface
type fakeReadServer struct {
	buffer    *bytes.Buffer
	sendCount int
	grpc.ServerStream
}

func makeFakeReadServer() *fakeReadServer {
	return &fakeReadServer{
		buffer: new(bytes.Buffer),
	}
}

func (s *fakeReadServer) Send(r *bytestream.ReadResponse) error {
	s.buffer.Write(r.GetData())
	s.sendCount++
	return nil
}

func (s *fakeReadServer) reset() {
	s.buffer.Reset()
	s.sendCount = 0
}

// Write server that fakes receiving specified data from a client.
// One instance will fake-write one piece of data identified by Digest(hash, size),
// []byte data. Chunks supports Recv'ing data in chunks to exercise the CAS server.
//	recvCount - tracks times the server invokes Recv for verification
//	committedSize - tracks the total data len Recv'd by the server for verification
// Implements bytestream.ByteStream_WriteServer interface
type fakeWriteServer struct {
	resourceName  string
	data          []byte
	recvChunks    int
	recvCount     int
	offset        int64
	committedSize int64
	grpc.ServerStream
}

func makeFakeWriteServer(hash string, size int64, data []byte, chunks int) *fakeWriteServer {
	uid, _ := uuid.NewV4()
	return &fakeWriteServer{
		resourceName:  fmt.Sprintf("uploads/%s/blobs/%s/%d", uid, hash, size),
		data:          data,
		recvChunks:    chunks,
		offset:        0,
		committedSize: 0,
	}
}

func (s *fakeWriteServer) SendAndClose(res *bytestream.WriteResponse) error {
	s.committedSize = res.GetCommittedSize()
	return nil
}

func (s *fakeWriteServer) Recv() (*bytestream.WriteRequest, error) {
	// Format a WriteRequest based on the chunks requested and the offset of what has been recvd
	chunkSize := int64(len(s.data) / s.recvChunks)
	if s.recvCount+1 >= s.recvChunks {
		chunkSize = int64(len(s.data)) - s.offset
	}
	finished := false
	if chunkSize+s.offset >= int64(len(s.data)) {
		finished = true
	}
	r := &bytestream.WriteRequest{
		ResourceName: s.resourceName,
		WriteOffset:  s.offset,
		FinishWrite:  finished,
		Data:         s.data[s.offset : s.offset+chunkSize],
	}
	s.offset = s.offset + chunkSize
	s.recvCount++
	return r, nil
}

// Serialize an ActionResult for placement in a Store for use in ActionCache testing
func getFakeActionResult() ([]byte, error) {
	d := &remoteexecution.Digest{Hash: testHash1, SizeBytes: testSize1}
	ar := &remoteexecution.ActionResult{
		OutputFiles: []*remoteexecution.OutputFile{
			&remoteexecution.OutputFile{Path: "/dir/file", Digest: d},
		},
		OutputDirectories: []*remoteexecution.OutputDirectory{
			&remoteexecution.OutputDirectory{Path: "/dir", TreeDigest: d},
		},
		ExitCode:     int32(12),
		StdoutDigest: d,
		StderrDigest: d,
	}
	arAsBytes, err := proto.Marshal(ar)
	if err != nil {
		return nil, err
	}
	return arAsBytes, nil
}

// Fake GetTreeServer
// Implements ContentAddressableStorage_GetTreeServer interface
type fakeGetTreeServer struct {
	grpc.ServerStream
}

func (s *fakeGetTreeServer) Send(*remoteexecution.GetTreeResponse) error {
	return nil
}
