package trie
import (
	"bytes"
	"testing"
	"github.com/DEL-ORG/del/common"
	"github.com/DEL-ORG/del/ethdb"
)
func makeTestTrie() (*Database, *Trie, map[string][]byte) {
	diskdb, _ := ethdb.NewMemDatabase()
	triedb := NewDatabase(diskdb)
	trie, _ := New(common.Hash{}, triedb)
	content := make(map[string][]byte)
	for i := byte(0); i < 255; i++ {
		key, val := common.LeftPadBytes([]byte{1, i}, 32), []byte{i}
		content[string(key)] = val
		trie.Update(key, val)
		key, val = common.LeftPadBytes([]byte{2, i}, 32), []byte{i}
		content[string(key)] = val
		trie.Update(key, val)
		for j := byte(3); j < 13; j++ {
			key, val = common.LeftPadBytes([]byte{j, i}, 32), []byte{j, i}
			content[string(key)] = val
			trie.Update(key, val)
		}
	}
	trie.Commit(nil)
	return triedb, trie, content
}
func checkTrieContents(t *testing.T, db *Database, root []byte, content map[string][]byte) {
	trie, err := New(common.BytesToHash(root), db)
	if err != nil {
		t.Fatalf("failed to create trie at %x: %v", root, err)
	}
	if err := checkTrieConsistency(db, common.BytesToHash(root)); err != nil {
		t.Fatalf("inconsistent trie at %x: %v", root, err)
	}
	for key, val := range content {
		if have := trie.Get([]byte(key)); !bytes.Equal(have, val) {
			t.Errorf("entry %x: content mismatch: have %x, want %x", key, have, val)
		}
	}
}
func checkTrieConsistency(db *Database, root common.Hash) error {
	trie, err := New(root, db)
	if err != nil {
		return nil 
	}
	it := trie.NodeIterator(nil)
	for it.Next(true) {
	}
	return it.Error()
}
func TestEmptyTrieSync(t *testing.T) {
	diskdbA, _ := ethdb.NewMemDatabase()
	triedbA := NewDatabase(diskdbA)
	diskdbB, _ := ethdb.NewMemDatabase()
	triedbB := NewDatabase(diskdbB)
	emptyA, _ := New(common.Hash{}, triedbA)
	emptyB, _ := New(emptyRoot, triedbB)
	for i, trie := range []*Trie{emptyA, emptyB} {
		diskdb, _ := ethdb.NewMemDatabase()
		if req := NewTrieSync(trie.Hash(), diskdb, nil).Missing(1); len(req) != 0 {
			t.Errorf("test %d: content requested for empty trie: %v", i, req)
		}
	}
}
func TestIterativeTrieSyncIndividual(t *testing.T) { testIterativeTrieSync(t, 1) }
func TestIterativeTrieSyncBatched(t *testing.T)    { testIterativeTrieSync(t, 100) }
func testIterativeTrieSync(t *testing.T, batch int) {
	srcDb, srcTrie, srcData := makeTestTrie()
	diskdb, _ := ethdb.NewMemDatabase()
	triedb := NewDatabase(diskdb)
	sched := NewTrieSync(srcTrie.Hash(), diskdb, nil)
	queue := append([]common.Hash{}, sched.Missing(batch)...)
	for len(queue) > 0 {
		results := make([]SyncResult, len(queue))
		for i, hash := range queue {
			data, err := srcDb.Node(hash)
			if err != nil {
				t.Fatalf("failed to retrieve node data for %x: %v", hash, err)
			}
			results[i] = SyncResult{hash, data}
		}
		if _, index, err := sched.Process(results); err != nil {
			t.Fatalf("failed to process result #%d: %v", index, err)
		}
		if index, err := sched.Commit(diskdb); err != nil {
			t.Fatalf("failed to commit data #%d: %v", index, err)
		}
		queue = append(queue[:0], sched.Missing(batch)...)
	}
	checkTrieContents(t, triedb, srcTrie.Root(), srcData)
}
func TestIterativeDelayedTrieSync(t *testing.T) {
	srcDb, srcTrie, srcData := makeTestTrie()
	diskdb, _ := ethdb.NewMemDatabase()
	triedb := NewDatabase(diskdb)
	sched := NewTrieSync(srcTrie.Hash(), diskdb, nil)
	queue := append([]common.Hash{}, sched.Missing(10000)...)
	for len(queue) > 0 {
		results := make([]SyncResult, len(queue)/2+1)
		for i, hash := range queue[:len(results)] {
			data, err := srcDb.Node(hash)
			if err != nil {
				t.Fatalf("failed to retrieve node data for %x: %v", hash, err)
			}
			results[i] = SyncResult{hash, data}
		}
		if _, index, err := sched.Process(results); err != nil {
			t.Fatalf("failed to process result #%d: %v", index, err)
		}
		if index, err := sched.Commit(diskdb); err != nil {
			t.Fatalf("failed to commit data #%d: %v", index, err)
		}
		queue = append(queue[len(results):], sched.Missing(10000)...)
	}
	checkTrieContents(t, triedb, srcTrie.Root(), srcData)
}
func TestIterativeRandomTrieSyncIndividual(t *testing.T) { testIterativeRandomTrieSync(t, 1) }
func TestIterativeRandomTrieSyncBatched(t *testing.T)    { testIterativeRandomTrieSync(t, 100) }
func testIterativeRandomTrieSync(t *testing.T, batch int) {
	srcDb, srcTrie, srcData := makeTestTrie()
	diskdb, _ := ethdb.NewMemDatabase()
	triedb := NewDatabase(diskdb)
	sched := NewTrieSync(srcTrie.Hash(), diskdb, nil)
	queue := make(map[common.Hash]struct{})
	for _, hash := range sched.Missing(batch) {
		queue[hash] = struct{}{}
	}
	for len(queue) > 0 {
		results := make([]SyncResult, 0, len(queue))
		for hash := range queue {
			data, err := srcDb.Node(hash)
			if err != nil {
				t.Fatalf("failed to retrieve node data for %x: %v", hash, err)
			}
			results = append(results, SyncResult{hash, data})
		}
		if _, index, err := sched.Process(results); err != nil {
			t.Fatalf("failed to process result #%d: %v", index, err)
		}
		if index, err := sched.Commit(diskdb); err != nil {
			t.Fatalf("failed to commit data #%d: %v", index, err)
		}
		queue = make(map[common.Hash]struct{})
		for _, hash := range sched.Missing(batch) {
			queue[hash] = struct{}{}
		}
	}
	checkTrieContents(t, triedb, srcTrie.Root(), srcData)
}
func TestIterativeRandomDelayedTrieSync(t *testing.T) {
	srcDb, srcTrie, srcData := makeTestTrie()
	diskdb, _ := ethdb.NewMemDatabase()
	triedb := NewDatabase(diskdb)
	sched := NewTrieSync(srcTrie.Hash(), diskdb, nil)
	queue := make(map[common.Hash]struct{})
	for _, hash := range sched.Missing(10000) {
		queue[hash] = struct{}{}
	}
	for len(queue) > 0 {
		results := make([]SyncResult, 0, len(queue)/2+1)
		for hash := range queue {
			data, err := srcDb.Node(hash)
			if err != nil {
				t.Fatalf("failed to retrieve node data for %x: %v", hash, err)
			}
			results = append(results, SyncResult{hash, data})
			if len(results) >= cap(results) {
				break
			}
		}
		if _, index, err := sched.Process(results); err != nil {
			t.Fatalf("failed to process result #%d: %v", index, err)
		}
		if index, err := sched.Commit(diskdb); err != nil {
			t.Fatalf("failed to commit data #%d: %v", index, err)
		}
		for _, result := range results {
			delete(queue, result.Hash)
		}
		for _, hash := range sched.Missing(10000) {
			queue[hash] = struct{}{}
		}
	}
	checkTrieContents(t, triedb, srcTrie.Root(), srcData)
}
func TestDuplicateAvoidanceTrieSync(t *testing.T) {
	srcDb, srcTrie, srcData := makeTestTrie()
	diskdb, _ := ethdb.NewMemDatabase()
	triedb := NewDatabase(diskdb)
	sched := NewTrieSync(srcTrie.Hash(), diskdb, nil)
	queue := append([]common.Hash{}, sched.Missing(0)...)
	requested := make(map[common.Hash]struct{})
	for len(queue) > 0 {
		results := make([]SyncResult, len(queue))
		for i, hash := range queue {
			data, err := srcDb.Node(hash)
			if err != nil {
				t.Fatalf("failed to retrieve node data for %x: %v", hash, err)
			}
			if _, ok := requested[hash]; ok {
				t.Errorf("hash %x already requested once", hash)
			}
			requested[hash] = struct{}{}
			results[i] = SyncResult{hash, data}
		}
		if _, index, err := sched.Process(results); err != nil {
			t.Fatalf("failed to process result #%d: %v", index, err)
		}
		if index, err := sched.Commit(diskdb); err != nil {
			t.Fatalf("failed to commit data #%d: %v", index, err)
		}
		queue = append(queue[:0], sched.Missing(0)...)
	}
	checkTrieContents(t, triedb, srcTrie.Root(), srcData)
}
func TestIncompleteTrieSync(t *testing.T) {
	srcDb, srcTrie, _ := makeTestTrie()
	diskdb, _ := ethdb.NewMemDatabase()
	triedb := NewDatabase(diskdb)
	sched := NewTrieSync(srcTrie.Hash(), diskdb, nil)
	added := []common.Hash{}
	queue := append([]common.Hash{}, sched.Missing(1)...)
	for len(queue) > 0 {
		results := make([]SyncResult, len(queue))
		for i, hash := range queue {
			data, err := srcDb.Node(hash)
			if err != nil {
				t.Fatalf("failed to retrieve node data for %x: %v", hash, err)
			}
			results[i] = SyncResult{hash, data}
		}
		if _, index, err := sched.Process(results); err != nil {
			t.Fatalf("failed to process result #%d: %v", index, err)
		}
		if index, err := sched.Commit(diskdb); err != nil {
			t.Fatalf("failed to commit data #%d: %v", index, err)
		}
		for _, result := range results {
			added = append(added, result.Hash)
		}
		for _, root := range added {
			if err := checkTrieConsistency(triedb, root); err != nil {
				t.Fatalf("trie inconsistent: %v", err)
			}
		}
		queue = append(queue[:0], sched.Missing(1)...)
	}
	for _, node := range added[1:] {
		key := node.Bytes()
		value, _ := diskdb.Get(key)
		diskdb.Delete(key)
		if err := checkTrieConsistency(triedb, added[0]); err == nil {
			t.Fatalf("trie inconsistency not caught, missing: %x", key)
		}
		diskdb.Put(key, value)
	}
}
