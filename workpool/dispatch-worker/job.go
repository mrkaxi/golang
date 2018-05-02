package main

import (
	"fmt"
	"time"
)

var JobQueue = make(chan Job, MaxQueue)

var count = int64(0)

//var JobQueue chan Job

type Payload struct {
	Id int32
}

// Job represents the job to be run
type Job struct {
	Payload *Payload
}

// A buffered channel that we can send work requests on.

func (self *Payload) Upload() {
	time.Sleep(5 * time.Second)
	//fmt.Println("id:", self.Id)
	count++
	fmt.Println("JobQueue", int32(len(JobQueue)), count)
}
