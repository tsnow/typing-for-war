package main

import "encoding/json"
import "fmt"
import "io/ioutil"
import "path/filepath"
import "os"
import "testing"

type testAttempt struct {
	Objective string
	Attempt   string
}

func testfile(dir string, name string) string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return filepath.Join(wd, dir, name)
}
func StoreTestCase(name string, obj string, attempt string) {
	out, err := json.Marshal(testAttempt{
		Objective: obj,
		Attempt:   attempt,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(testfile("corpus", name))
	ioutil.WriteFile(testfile("corpus", name), out, 0777)
	out, err = json.Marshal(GoodBadLeft(obj, attempt))
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile(testfile("expected", name), out, 0777)
}

func LoadTestCase(name string) (testAttempt, *playState) {
	buf, err := ioutil.ReadFile(testfile("corpus", name))
	if err != nil {
		panic(err)
	}
	var attempt testAttempt
	err = json.Unmarshal(buf, &attempt)
	if err != nil {
		panic(err)
	}
	buf, err = ioutil.ReadFile(testfile("expected", name))
	if err != nil {
		fmt.Printf("error=\"%s\" filename=\"%s\"", err, testfile("expected", name))
		return attempt, nil
	}
	var playState playState
	err = json.Unmarshal(buf, &playState)
	return attempt, &playState
}

/*
func TestMakeTests(t *testing.T){
	StoreTestCase("0_empty","", "")
	StoreTestCase("1_empty_extra","", "1")
	StoreTestCase("2_initial_game","1", "")
	StoreTestCase("3_finished_game","1", "1")
	StoreTestCase("4_wrong_char","1", "2")
	StoreTestCase("5_half_game","12", "1")
	StoreTestCase("6_missed_a_char", "bobwehadababyitsaboy", "bobwehadababyts")
}
*/
func verifyTestCase(name string) error {
	testcase, msg := LoadTestCase(name)
	actual_msg := GoodBadLeft(testcase.Objective, testcase.Attempt)
	if msg == nil {
		return fmt.Errorf("GoodBadLeft "+name+": no expected/ file got %q", actual_msg)
	}
	if actual_msg != *msg {
		return fmt.Errorf("GoodBadLeft "+name+": expected %q got %q", *msg, actual_msg)
	}
	return nil
}
func TestGoodBadLeft(t *testing.T) {
	cases := []string{
		"0_empty",
		"1_empty_extra",
		"2_initial_game",
		"3_finished_game",
		"4_wrong_char",
		"5_half_game",
		"6_missed_a_char",
	}
	for _, i := range cases {
		fail := verifyTestCase(i)
		if fail != nil {
			t.Error(fail)
		}
	}
	var msg, actual_msg playState
	msg = playState{"", "", ""}
	actual_msg = GoodBadLeft("", "")
	if msg != actual_msg {

	}
	msg = playState{"", "1", ""}
	actual_msg = GoodBadLeft("", "1")
	if msg != actual_msg {
		t.Errorf("GoodBadLeft ',1': expected %q got %q", msg, actual_msg)
	}
	msg = playState{"", "", "1"}
	actual_msg = GoodBadLeft("1", "")
	if msg != actual_msg {
		t.Errorf("GoodBadLeft '1,': expected %q got %q", msg, actual_msg)
	}
	msg = playState{"1", "", ""}
	actual_msg = GoodBadLeft("1", "1")
	if msg != actual_msg {
		t.Errorf("GoodBadLeft '1,1': expected %q got %q", msg, actual_msg)
	}
	msg = playState{"", "2", "1"}
	actual_msg = GoodBadLeft("1", "2")
	if msg != actual_msg {
		t.Errorf("GoodBadLeft '1,2': expected %q got %q", msg, actual_msg)
	}
	msg = playState{"1", "", "2"}
	actual_msg = GoodBadLeft("12", "1")
	if msg != actual_msg {
		t.Errorf("GoodBadLeft '12,1': expected %q got %q", msg, actual_msg)
	}
	msg = playState{"bobwehadababy", "ts", "itsaboy"}
	actual_msg = GoodBadLeft("bobwehadababyitsaboy", "bobwehadababyts")
	if msg != actual_msg {
		t.Errorf("GoodBadLeft 'bobwehadababyitsaboy,bobwehadababyts': expected %q got %q", msg, actual_msg)
	}
}

/*
func TestFuzz0d7c23e186a4b1061e0d39d9e3a26d65372e1800_4(t *testing.T){
    gbl := tfw.GoodBadLeft("468111839T96a789a3bc0045c8\x80\xff\xff\xff42c7d1bd998f5\x8a4495X9b4", "468111839T96a789a3bc0045c8\x80\xff\xff\xff42c7d1bd998f57")
    t.Error(gbl)
}

func TestFuzz1719ef6fabace958e726b5fec0841ec681a87941_1(t *testing.T){
    gbl := tfw.GoodBadLeft("012345678901234567890123456789012346789012345678", "90123456789012345678901234\xbf\xbd89")
    t.Error(gbl)
}

func TestFuzz3d1299ac3a885e69d21f47c3865f2cdca51eab6b_2(t *testing.T){
    gbl := tfw.GoodBadLeft("\xef\xbf", "\xef")
    t.Error(gbl)
}

func TestFuzz51a08443aea76c712e712ebf0659c41de3165bbb_1(t *testing.T){
    gbl := tfw.GoodBadLeft("11839T96a789a3bc0045c8a5fb42c7d1bd998f54449579b44", "6817afbd17273e662c97\xff\x0072995ef42640c550b9013fad0761353c7086a272c24088be94769fd1665")
    t.Error(gbl)
}

func TestFuzz7cdb9a30e800f4078b7a82485e650bfd403882bd_2(t *testing.T){
    gbl := tfw.GoodBadLeft("468111839T96a789a3bc0045c8a5fb42c7d1bd998f54449579b4", "46817")
    t.Error(gbl)
}

func TestFuzzae533bd603a632f22ffff519f377c3fbe737ee3e_3(t *testing.T){
    gbl := tfw.GoodBadLeft("11839T96a789a3bc0�^f\r5fb42c7d1bd9984449579b4468", "17")
    t.Error(gbl)
}

func TestFuzzb5f078fcf6ad1b7141ed045107ebc32840cd7cfa_2(t *testing.T){
    gbl := tfw.GoodBadLeft("\xbd\xbfｿｿ\xed\xbd\xbfｿｿ", "\xef")
    t.Error(gbl)
}

func TestFuzzb9fd949e2778beb7f87fce4beb29a0c326f6afd2_1(t *testing.T){
    gbl := tfw.GoodBadLeft("�!�", "\x05")
    t.Error(gbl)
}

func TestFuzzccf6ec39c42d8442e42d80ee6ca99da859e84852_2(t *testing.T){
    gbl := tfw.GoodBadLeft("11839T96a789a3bc0045c8a5fb42c7d1bd9984449579b4468", "17")
    t.Error(gbl)
}



func TestFuzze4a5c9ec08c116b55a753d4556a1f8a36e6675ed_3(t *testing.T){
    gbl := tfw.GoodBadLeft("468111839T96a789a3bc0045c8a5fb42c7d1bd998f54449579b4", "46811184")
    t.Error(gbl)
}

func TestFuzzea24535cc5faf8d9ca733cc0239235e7b8071202_3(t *testing.T){
    gbl := tfw.GoodBadLeft("468111839T96a789a3bc0045c8\x80\xff\xff\xff42c7d1bd998f544495X9b4", "468111839T96a789a3b\xff")
    t.Error(gbl)
}

func TestFuzzef90a7bc3fd93b3f5821e687bd4d1951fecd55a3_3(t *testing.T){
    gbl := tfw.GoodBadLeft("A1839T96a789a3bc0045c-51633419fb42c7d1bd998f54449579b446817afbd17", "273e662c97\xff\x0072995ef42\xfc40c550b9013fad0761353c7086a272c24088be94769")
    t.Error(gbl)
}

func TestFuzzefe88c9c3a772207bd047d82293e0a1469fc395d_1(t *testing.T){
    gbl := tfw.GoodBadLeft("\xbd\xbfｿｿｿ", "\xef")
    t.Error(gbl)
}

func TestFuzzfe564f789dc654a770ed03187c2200c93ee7be3d_1(t *testing.T){
    gbl := tfw.GoodBadLeft("D", "i")
    t.Error(gbl)
}

*/
