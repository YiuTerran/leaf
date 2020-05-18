package tz

import (
	"testing"
	"time"
)

func TestSchedule(t *testing.T) {
	ch := make(chan struct{}, 1)
	//Schedule(func() {
	//	time.Sleep(2 * time.Second)
	//	t.Logf("s:%s\n", time.Now())
	//}, 5*time.Second, ch)
	//Cycle(func() {
	//	time.Sleep(2 * time.Second)
	//	t.Logf("c:%s\n", time.Now())
	//}, 5*time.Second, ch)
	d1 := 1 * time.Second
	d2 := 1 * time.Second
	//3->4->5->6->7
	DynamicSchedule(func() {
		t.Logf("ds:%s>\n", time.Now().Format(FullFormat))
		time.Sleep(1 * time.Second)
		d1 += 1 * time.Second
	}, &d1, ch)
	//1->4->5->6->7->8->9
	DynamicCycle(func() {
		t.Logf("dc:%s<\n", time.Now().Format(FullFormat))
		time.Sleep(1 * time.Second)
		d2 += 1 * time.Second
	}, &d2, ch)
	time.Sleep(60 * time.Second)
}
