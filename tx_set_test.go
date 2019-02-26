package nutsdb

import (
	"io/ioutil"
	"os"
	"testing"
)

func InitForSet() {
	fileDir := "/tmp/nutsdbtestsettx"
	files, _ := ioutil.ReadDir(fileDir)
	for _, f := range files {
		name := f.Name()
		if name != "" {
			err := os.Remove(fileDir + "/" + name)
			if err != nil {
				panic(err)
			}
		}
	}

	opt = DefaultOptions
	opt.Dir = fileDir
	opt.SegmentSize = 8 * 1024
	return
}

func TestTx_SAdd(t *testing.T) {
	InitForSet()
	db, err = Open(opt)
	if err != nil {
		t.Fatal(err)
	}

	//write tx begin
	tx, err := db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}

	bucket := "bucket1"
	key := []byte("key1")
	val1 := []byte("val1")
	val2 := []byte("val2")

	if err := tx.SAdd(bucket, []byte(""), val1, val2); err == nil {
		t.Error("TestTx_SAdd err")
		t.Fatal(err)
	}

	if err := tx.SAdd(bucket, key, val1, val2); err != nil {
		err = tx.Rollback()
		t.Fatal(err)
	} else {
		tx.Commit()
	}

	//read tx
	tx, err = db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}

	if ok, err := tx.SAreMembers(bucket, key, val1, val2); err != nil {
		tx.Rollback()
		t.Fatal(err)
	} else {
		tx.Commit()
		if !ok {
			t.Error("TestTx_SAdd err")
		}
	}

}

func TestTx_SRem(t *testing.T) {
	InitForSet()
	db, err = Open(opt)

	//write tx begin
	tx, err := db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}

	bucket := "bucket2"
	key := []byte("key1")
	val1 := []byte("one")
	val2 := []byte("two")
	val3 := []byte("three")

	if err := tx.SAdd(bucket, key, val1, val2, val3); err != nil {
		err = tx.Rollback()
		t.Fatal(err)
	} else {
		tx.Commit()
	}

	//write tx begin
	tx, err = db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}

	if err = tx.SRem(bucket, key, val3); err != nil {
		tx.Rollback()
		t.Fatal(err)
	} else {
		tx.Commit()
	}

	//read tx
	tx, err = db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}

	if ok, err := tx.SAreMembers(bucket, key, val1, val2); err != nil {
		err = tx.Rollback()
	} else {
		tx.Commit()
		if !ok {
			t.Error("TestTx_SRem err")
		}
	}

	tx, err = db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := tx.SIsMember(bucket, key, val3)
	if err == nil && ok {
		t.Error("TestTx_SRem err")
	}

	tx.Rollback()
}

func TestTx_SMembers(t *testing.T) {
	InitForSet()
	db, err = Open(opt)

	//write tx begin
	tx, err := db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}

	bucket := "bucket3"
	key := []byte("key1")
	val1 := []byte("Hello")
	val2 := []byte("World")

	if err := tx.SAdd(bucket, key, val1, val2); err != nil {
		err = tx.Rollback()
		t.Fatal(err)
	} else {
		tx.Commit()
	}

	//read tx
	tx, err = db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}

	if list, err := tx.SMembers(bucket, key); err != nil {
		err = tx.Rollback()
	} else {
		if len(list) != 2 {
			t.Error("TestTx_SMembers err")
		}

		if ok, _ := tx.SIsMember(bucket, key, []byte("Hello")); !ok {
			t.Error("TestTx_SMembers err")
		}

		if ok, _ := tx.SIsMember(bucket, key, []byte("World")); !ok {
			t.Error("TestTx_SMembers err")
		}

		list, err := tx.SMembers("fake_bucket", key)
		if len(list) > 0 || err == nil {
			t.Error("TestTx_SMembers err")
		}

		tx.Commit()

		list, err = tx.SMembers(bucket, key)
		if len(list) == 2 || err == nil {
			t.Error("TestTx_SMembers err")
		}
	}
}

func TestTx_SCard(t *testing.T) {
	InitForSet()
	db, err = Open(opt)

	//write tx begin
	tx, err := db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}

	bucket := "bucket4"
	key := []byte("key1")
	val1 := []byte("1")
	val2 := []byte("2")
	val3 := []byte("3")

	if err := tx.SAdd(bucket, key, val1, val2, val3); err != nil {
		err = tx.Rollback()
		t.Fatal(err)
	} else {
		tx.Commit()
	}

	tx, err = db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}

	if num, err := tx.SCard(bucket, key); num != 3 && err != nil {
		tx.Rollback()
		t.Error("TestSet_SCard err")
	}

	if num, err := tx.SCard("key_fake", key); err == nil {
		tx.Rollback()
		t.Error("TestSet_SCard err")
	} else {
		if num != 0 {
			tx.Rollback()
			t.Error("TestSet_SCard err")
		}
		tx.Commit()

		num, err = tx.SCard(bucket, key)
		if num > 0 || err == nil {
			t.Error("TestTx_SCard err")
		}
	}
}

func TestTx_SDiffByOneBucket(t *testing.T) {
	InitForSet()
	db, err = Open(opt)

	tx, err := db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}

	bucket := "bucket5"
	key1 := []byte("mySet1")
	key2 := []byte("mySet2")

	if err := tx.SAdd(bucket, key1, []byte("a"), []byte("b"), []byte("c")); err != nil {
		tx.Rollback()
		t.Fatal(err)
	} else {
		err = tx.SAdd(bucket, key2, []byte("c"), []byte("d"), []byte("e"))
		if err != nil {
			tx.Rollback()
			t.Fatal(err)
		}

		tx.Commit()
	}

	tx, err = db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}

	list, err := tx.SDiffByOneBucket(bucket, key1, key2)
	if err != nil {
		tx.Rollback()
		t.Fatal(err)
	} else {

		list, err = tx.SDiffByOneBucket("fake_bucket", key1, key2)
		if err == nil || list != nil {
			t.Error("TestTx_SDiffByOneBucket err")
		}

		tx.Commit()
	}

	tx, err = db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}

	key3 := []byte("mySet3")
	if err = tx.SAdd(bucket, key3, []byte("a"), []byte("b")); err != nil {
		tx.Rollback()
		t.Fatal(err)
	} else {
		tx.Commit()
	}

	tx, err = db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range list {
		if ok, _ := tx.SIsMember(bucket, key3, item); !ok {
			t.Error("TestTx_SDiffByOneBucket err")
		}
	}

	tx.Commit()

	list, err = tx.SDiffByOneBucket(bucket, key1, key2)
	if err == nil || list != nil {
		t.Error("TestTx_SDiffByOneBucket err")
	}
}

func TestTx_SDiffByTwoBuckets(t *testing.T) {
	InitForSet()
	db, err = Open(opt)

	tx, err := db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}

	bucket1 := "bucket6"
	key1 := []byte("mySet1")
	if err := tx.SAdd(bucket1, key1, []byte("a"), []byte("b"), []byte("c")); err != nil {
		tx.Rollback()
		t.Fatal(err)
	}

	bucket2 := "bucket7"
	key2 := []byte("mySet2")
	if err := tx.SAdd(bucket2, key2, []byte("c"), []byte("d"), []byte("e")); err != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	tx.Commit()

	tx, err = db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}

	list, err := tx.SDiffByTwoBuckets(bucket1, key1, bucket2, key2)
	if err != nil {
		tx.Rollback()
		t.Fatal(err)
	} else {

		list, err = tx.SDiffByTwoBuckets("fake_bucket1", key1, bucket2, key2)
		if err == nil || list != nil {
			t.Error("TestTx_SDiffByTwoBuckets err")
		}

		list, err = tx.SDiffByTwoBuckets(bucket1, key1, "fake_bucket2", key2)
		if err == nil || list != nil {
			t.Error("TestTx_SDiffByTwoBuckets err")
		}

		tx.Commit()
	}

	tx, err = db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}

	key3 := []byte("mySet3")
	bucket := "bucket8"
	if err = tx.SAdd(bucket, key3, []byte("a"), []byte("b")); err != nil {
		tx.Rollback()
		t.Fatal(err)
	} else {
		tx.Commit()
	}

	tx, err = db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range list {
		if ok, _ := tx.SIsMember(bucket, key3, item); !ok {
			t.Error("TestTx_SDiffByTwoBuckets err")
		}
	}

	tx.Commit()

	list, err = tx.SDiffByTwoBuckets(bucket1, key1, bucket2, key2)
	if err == nil || list != nil {
		t.Error("TestTx_SDiffByTwoBuckets err")
	}
}

func TestTx_SPop(t *testing.T) {
	InitForSet()
	db, err = Open(opt)

	tx, err := db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}

	bucket := "bucket9"
	key := []byte("mySet")
	if err = tx.SAdd(bucket, key, []byte("one"), []byte("two"), []byte("three")); err != nil {
		tx.Rollback()
		t.Fatal(err)
	} else {
		tx.Commit()
	}

	tx, err = db.Begin(false)
	num, err := tx.SCard(bucket, key)
	if num != 3 {
		tx.Rollback()
		t.Fatal("TestTx_SPop err")
	} else {
		tx.Commit()
	}

	tx, err = db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}

	item, err := tx.SPop(bucket, key)
	if err != nil || item == nil {
		tx.Rollback()
		t.Fatal(err)
	}

	item, err = tx.SPop("fake_bucket", key)
	if err == nil {
		tx.Rollback()
		t.Fatal("TestTx_SPop err")
	}

	ok, err := tx.SIsMember(bucket, key, item)
	if ok && err == nil {
		t.Error("TestTx_SPop err")
	}

	tx.Commit()

	tx, err = db.Begin(false)
	num, err = tx.SCard(bucket, key)
	if num != 2 {
		tx.Rollback()
		t.Fatal("TestTx_SPop err")
	} else {
		tx.Commit()
	}


	item, err = tx.SPop(bucket, key)
	if err == nil || item != nil {
		t.Fatal(err)
	}
}

func TestTx_SMoveByOneBucket(t *testing.T) {
	InitForSet()
	db, err = Open(opt)

	tx, err := db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}

	bucket := "bucket10"

	key1 := []byte("mySet1")
	err = tx.SAdd(bucket, key1, []byte("one"), []byte("two"))
	if err != nil {
		tx.Rollback()
		t.Fatal(err)
	}

	key2 := []byte("mySet2")
	tx.SAdd(bucket, key2, []byte("three"))
	if err != nil {
		tx.Rollback()
		t.Fatal(err)
	}

	tx.Commit()

	tx, err = db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := tx.SMoveByOneBucket(bucket, key1, key2, []byte("two"))
	if !ok {
		t.Error("TestTx_SMoveByOneBucket err")
	}
	if err != nil {
		tx.Rollback()
		t.Fatal(err)
	} else {
		ok, err = tx.SMoveByOneBucket("fake_bucket", key1, key2, []byte("two"))
		if ok || err == nil {
			t.Error("TestTx_SMoveByOneBucket err")
		}
		tx.Commit()
	}

	tx, err = db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}

	ok, err = tx.SIsMember(bucket, key1, []byte("two"))
	if ok || err == nil {
		t.Error("TestTx_SMoveByOneBucket err")
	}

	tx.Commit()

	tx, err = db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}

	ok, err = tx.SIsMember(bucket, key2, []byte("two"))
	if !ok || err != nil {
		t.Error("TestTx_SMoveByOneBucket err")
	}

	tx.Commit()

	ok, err = tx.SMoveByOneBucket(bucket, key1, key2, []byte("two"))
	if ok || err == nil {
		t.Error("TestTx_SMoveByOneBucket err")
	}
}

func TestTx_SMoveByTwoBuckets(t *testing.T) {
	InitForSet()
	db, err = Open(opt)

	tx, err := db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}

	bucket1 := "bucket11"
	key1 := []byte("mySet1")
	bucket2 := "bucket12"
	key2 := []byte("mySet2")

	if err = tx.SAdd(bucket1, key1, []byte("one"), []byte("two")); err != nil {
		tx.Rollback()
		t.Fatal(err)
	}

	if err = tx.SAdd(bucket2, key2, []byte("three")); err != nil {
		tx.Rollback()
		t.Fatal(err)
	}

	tx.Commit()

	tx, err = db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := tx.SMoveByTwoBuckets(bucket1, []byte("fake_mySet1"), bucket2, key2, []byte("two"))
	if err == nil || ok {
		t.Error("TestTx_SMoveByTwoBuckets err")
	}

	ok, err = tx.SMoveByTwoBuckets(bucket1, key1, bucket2, []byte("fake_mySet2"), []byte("two"))
	if err == nil || ok {
		t.Error("TestTx_SMoveByTwoBuckets err")
	}

	ok, err = tx.SMoveByTwoBuckets(bucket1, key1, bucket2, key2, []byte("two"))
	if err != nil || !ok {
		t.Error("TestTx_SMoveByTwoBuckets err")
	}

	ok, err = tx.SMoveByTwoBuckets("fake_bucket1", key1, bucket2, key2, []byte("two"))
	if err == nil || ok {
		t.Error("TestTx_SMoveByTwoBuckets err")
	}

	ok, err = tx.SMoveByTwoBuckets(bucket1, key1, "fake_bucket2", key2, []byte("two"))
	if err == nil || ok {
		t.Error("TestTx_SMoveByTwoBuckets err")
	}

	ok, err = tx.SMoveByTwoBuckets("fake_bucket1", key1, "fake_bucket2", key2, []byte("two"))
	if err == nil || ok {
		t.Error("TestTx_SMoveByTwoBuckets err")
	}

	tx.Commit()

	tx, err = db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}

	ok, err = tx.SIsMember(bucket1, key1, []byte("two"))
	if ok || err == nil {
		t.Error("TestTx_SMoveByOneBucket err")
	}

	tx.Commit()

	tx, err = db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}

	ok, err = tx.SIsMember(bucket2, key2, []byte("two"))
	if !ok || err != nil {
		t.Error("TestTx_SMoveByTwoBuckets err")
	}

	tx.Commit()

	ok, err = tx.SMoveByTwoBuckets(bucket1, key1, bucket2, key2, []byte("two"))
	if ok || err == nil {
		t.Error("TestTx_SMoveByTwoBuckets err")
	}
}

func TestTx_SUnionByOneBucket(t *testing.T) {
	InitForSet()
	db, err = Open(opt)

	tx, err := db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}

	bucket := "bucket13"

	key1 := []byte("mySet1")
	err = tx.SAdd(bucket, key1, []byte("one"), []byte("two"))
	if err != nil {
		tx.Rollback()
		t.Fatal(err)
	}

	key2 := []byte("mySet2")
	tx.SAdd(bucket, key2, []byte("three"))
	if err != nil {
		tx.Rollback()
		t.Fatal(err)
	}

	key3 := []byte("mySet3")
	tx.SAdd(bucket, key3, []byte("one"), []byte("two"), []byte("three"))
	if err != nil {
		tx.Rollback()
		t.Fatal(err)
	}

	tx.Commit()

	tx, err = db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}

	list, err := tx.SUnionByOneBucket(bucket, key1, key2)

	if err != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	if len(list) != 3 {
		t.Error("TestTx_SUnionByOneBucket err")
	}

	for _, item := range list {
		if ok, _ := tx.SIsMember(bucket, key3, item); !ok {
			t.Error("TestTx_SUnionByOneBucket err")
		}
	}

	list, err = tx.SUnionByOneBucket("fake_bucket", key1, key2)
	if err == nil || list != nil {
		t.Error("TestTx_SUnionByOneBucket err")
	}

	tx.Commit()

	list, err = tx.SUnionByOneBucket(bucket, key1, key2)
	if list != nil || err == nil {
		t.Error("TestTx_SUnionByOneBucket err")
	}
}

func TestTx_SUnionByTwoBuckets(t *testing.T) {
	InitForSet()
	db, err = Open(opt)

	tx, err := db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}

	bucket1 := "bucket14"
	key1 := []byte("mySet1")
	bucket2 := "bucket15"
	key2 := []byte("mySet2")

	if err = tx.SAdd(bucket1, key1, []byte("one"), []byte("two")); err != nil {
		tx.Rollback()
		t.Fatal(err)
	}

	if err = tx.SAdd(bucket2, key2, []byte("three")); err != nil {
		tx.Rollback()
		t.Fatal(err)
	}

	tx.Commit()

	tx, err = db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}

	list, err := tx.SUnionByTwoBuckets(bucket1, key1, bucket2, key2)
	if err != nil {
		tx.Rollback()
		t.Fatal(err)
	}

	if len(list) != 3 {
		t.Error("TestTx_SUnionByTwoBuckets err")
	}

	list, err = tx.SUnionByTwoBuckets("fake_bucket1", key1, bucket2, key2)
	if list != nil || err == nil {
		t.Error("TestTx_SUnionByTwoBuckets err")
	}

	list, err = tx.SUnionByTwoBuckets(bucket1, key1, "fake_bucket2", key2)
	if list != nil || err == nil {
		t.Error("TestTx_SUnionByTwoBuckets err")
	}

	list, err = tx.SUnionByTwoBuckets(bucket1, []byte("fake_key1"), bucket2, key2)
	if list != nil || err == nil {
		t.Error("TestTx_SUnionByTwoBuckets err")
	}

	list, err = tx.SUnionByTwoBuckets(bucket1, key1, bucket2, []byte("fake_key2"))
	if list != nil || err == nil {
		t.Error("TestTx_SUnionByTwoBuckets err")
	}

	tx.Commit()

	list, err = tx.SUnionByTwoBuckets(bucket1, key1, bucket2, key2)
	if list != nil || err == nil {
		t.Error("TestTx_SUnionByTwoBuckets err")
	}
}

func TestTx_SHasKey(t *testing.T) {
	InitForSet()
	db, err = Open(opt)

	tx, err := db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}

	bucket := "bucket16"

	key1 := []byte("mySet1")
	err = tx.SAdd(bucket, key1, []byte("one"), []byte("two"))
	if err != nil {
		tx.Rollback()
		t.Fatal(err)
	}

	tx.Commit()

	tx, err = db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := tx.SHasKey(bucket, key1)
	if !ok || err != nil {
		t.Error("TestTx_SHasKey err")
	}

	ok, err = tx.SHasKey("fake_bucket", key1)
	if err == nil || ok {
		t.Error("TestTx_SHasKey err")
	}

	tx.Commit()

	ok, err = tx.SHasKey(bucket, key1)
	if err == nil || ok {
		t.Error("TestTx_SHasKey err")
	}
}

func TestTx_SIsMember(t *testing.T) {
	InitForSet()
	db, err = Open(opt)

	tx, err := db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}

	bucket := "bucket17"
	key := []byte("mySet")

	if err = tx.SAdd(bucket, key, []byte("Hello"), []byte("World")); err != nil {
		tx.Rollback()
		t.Fatal(err)
	}

	tx.Commit()

	tx, err = db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := tx.SIsMember(bucket, key, []byte("Hello"))
	if !ok || err != nil {
		t.Error("TestTx_SIsMember err")
	}

	ok, err = tx.SIsMember(bucket, key, []byte("World"))
	if !ok || err != nil {
		t.Error("TestTx_SIsMember err")
	}

	ok, err = tx.SIsMember(bucket, []byte("fake_key"), []byte("World"))
	if ok || err == nil {
		t.Error("TestTx_SIsMember err")
	}

	ok, err = tx.SIsMember(bucket, key, []byte("World2"))
	if ok || err == nil {
		t.Error("TestTx_SIsMember err")
	}

	ok, err = tx.SIsMember("fake_bucket", key, []byte("World"))
	if ok || err == nil {
		t.Error("TestTx_SIsMember err")
	}
	tx.Commit()

	ok, err = tx.SIsMember(bucket, key, []byte("World"))
	if ok || err == nil {
		t.Error("TestTx_SIsMember err")
	}
}

func TestTx_SAreMembers(t *testing.T) {
	InitForSet()
	db, err = Open(opt)

	tx, err := db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}

	bucket := "bucket18"
	key := []byte("mySet")

	if err = tx.SAdd(bucket, key, []byte("Hello"), []byte("World")); err != nil {
		tx.Rollback()
		t.Fatal(err)
	}

	tx.Commit()

	tx, err = db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := tx.SAreMembers(bucket, key, []byte("Hello"))
	if !ok || err != nil {
		t.Error("TestTx_SAreMembers err")
	}

	ok, err = tx.SAreMembers(bucket, key, []byte("World"))
	if !ok || err != nil {
		t.Error("TestTx_SAreMembers err")
	}

	ok, err = tx.SAreMembers(bucket, key, []byte("Hello"), []byte("World"))
	if !ok || err != nil {
		t.Error("TestTx_SAreMembers err")
	}

	ok, err = tx.SAreMembers(bucket, key, []byte("Hello2"), []byte("World"))
	if ok || err == nil {
		t.Error("TestTx_SAreMembers err")
	}

	ok, err = tx.SAreMembers("fake_bucket", key, []byte("Hello"), []byte("World"))
	if ok || err == nil {
		t.Error("TestTx_SAreMembers err")
	}

	tx.Commit()

	ok, err = tx.SAreMembers(bucket, key, []byte("Hello"), []byte("World"))
	if ok || err == nil {
		t.Error("TestTx_SAreMembers err")
	}
}
