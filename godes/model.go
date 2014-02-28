// Copyright 2013 Alex Goussiatiner. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package godes

import (
	"container/list"
	"fmt"
	//"sync"
	"time"
)

type Model struct {
	//mu                  sync.RWMutex
	activeRunner        RunnerInterface
	movingList          *list.List
	scheduledList       *list.List
	waitingList         *list.List
	waitingConditionMap map[int]RunnerInterface
	interruptedMap      map[int]RunnerInterface
	terminatedList      *list.List
	currentId           int
	controlChannel      chan int
	simulationActive    bool
	DEBUG               bool
}

//newModel initilizes the model
func newModel(verbose bool) *Model {

	var ball *Runner = NewRunner()
	ball.channel = make(chan int)
	ball.markTime = time.Now()
	ball.id = 0
	ball.state = RUNNER_STATE_ACTIVE //that is bypassing READY
	ball.priority = 100
	ball.setMarkTime(time.Now())
	var runner RunnerInterface = ball
	mdl := Model{activeRunner: runner, controlChannel: make(chan int), DEBUG: verbose, simulationActive: false}
	mdl.addToMovingList(runner)
	return &mdl
}

//advance model time
func (mdl *Model) advance(interval float64) bool {

	ch := mdl.activeRunner.getChannel()
	mdl.activeRunner.setMovingTime(stime + interval)
	mdl.activeRunner.setState(RUNNER_STATE_SCHEDULED)
	mdl.removeFromMovingList(mdl.activeRunner)
	mdl.addToSchedulledList(mdl.activeRunner)
	//restart control channel and freez
	mdl.controlChannel <- 100
	<-ch
	return true
}

//waitUntillDone waits untill all childs have returned
func (mdl *Model) waitUntillDone() {

	if mdl.activeRunner.GetId() != 0 {
		panic("waitUntillDone initiated for not main ball")
	}

	mdl.removeFromMovingList(mdl.activeRunner)
	mdl.controlChannel <- 100
	for {

		if !model.simulationActive {
			break
		} else {
			if mdl.DEBUG {
				fmt.Println("waiting", mdl.movingList.Len())
			}
			time.Sleep(time.Millisecond * simulationSecondScale)
		}
	}
}

//add adds a new runner to the model, adding it to the moving List
func (mdl *Model) add(runner RunnerInterface) bool {

	mdl.currentId++
	runner.setChannel(make(chan int))
	runner.setMovingTime(stime)
	runner.setId(mdl.currentId)
	runner.setState(RUNNER_STATE_READY)
	mdl.addToMovingList(runner)

	go func() {
		<-runner.getChannel()
		runner.setMarkTime(time.Now())
		runner.Run()
		if mdl.activeRunner == nil {
			panic("remove: activeRunner == nil")
		}
		mdl.removeFromMovingList(mdl.activeRunner)
		mdl.activeRunner.setState(RUNNER_STATE_TERMINATED)
		mdl.activeRunner = nil
		mdl.controlChannel <- 100
	}()
	return true

}

//interrupt interrupts the runner and removes it from the SchedulledList
func (mdl *Model) interrupt(runner RunnerInterface) {

	if runner.GetState() != RUNNER_STATE_SCHEDULED {
		panic("It is not  RUNNER_STATE_SCHEDULED")
	}
	mdl.removeFromSchedulledList(runner)
	runner.setState(RUNNER_STATE_INTERRUPTED)
	mdl.addToInterruptedMap(runner)

}

//resume resumes the execution of a runner with a new timestam = time of interruption + timeChange
func (mdl *Model) resume(runner RunnerInterface, timeChange float64) {
	if runner.GetState() != RUNNER_STATE_INTERRUPTED {
		panic("It is not  RUNNER_STATE_INTERRUPTED")
	}
	mdl.removeFromInterruptedMap(runner)
	runner.setState(RUNNER_STATE_SCHEDULED)
	runner.setMovingTime(runner.GetMovingTime() + timeChange)
	mdl.addToMovingList(runner)

}


func (mdl *Model) booleanControlWait(b *BooleanControl, val bool) {

	ch := mdl.activeRunner.getChannel()
	if mdl.activeRunner == nil {
		panic("booleanControlWait - no runner")
	}

	mdl.removeFromMovingList(mdl.activeRunner)

	mdl.activeRunner.setState(RUNNER_STATE_WAITING_COND)
	mdl.activeRunner.setWaitingForBool(val)
	mdl.activeRunner.setWaitingForBoolControl(b)

	mdl.addToWaitingConditionMap(mdl.activeRunner)
	mdl.controlChannel <- 100
	<-ch

}

func (mdl *Model) booleanControlWaitAndTimeout(b *BooleanControl, val bool, timeout float64) {

	ri := &TimeoutRunner{&Runner{}, mdl.activeRunner, timeout}
	AddRunner(ri)
	mdl.activeRunner.setWaitingForBoolControlTimeoutId(ri.GetId())
	mdl.booleanControlWait(b, val)

}

func (mdl *Model) booleanControlSet(b *BooleanControl) {
	ch := mdl.activeRunner.getChannel()
	if mdl.activeRunner == nil {
		panic("booleanControlSet - no runner")
	}
	mdl.controlChannel <- 100
	<-ch

}

func (mdl *Model) control() bool {

	if mdl.activeRunner == nil {
		panic("control: activeBall == nil")
	}

	go func() {
		var runner RunnerInterface
		for {
			<-mdl.controlChannel
			if mdl.waitingConditionMap != nil && len(mdl.waitingConditionMap) > 0 {
				for key, temp := range mdl.waitingConditionMap {
					if temp.getWaitingForBoolControl() == nil {
						panic("  no BoolControl")
					}
					if temp.getWaitingForBool() == temp.getWaitingForBoolControl().GetState() {
						temp.setState(RUNNER_STATE_READY)
						temp.setWaitingForBoolControl(nil)
						temp.setWaitingForBoolControlTimeoutId(-1)
						mdl.addToMovingList(temp)
						delete(mdl.waitingConditionMap, key)
						break
					}
				}
			}

			//finding new runner
			runner = nil
			if mdl.movingList != nil && mdl.movingList.Len() > 0 {
				runner = mdl.getFromMovingList()
			}
			if runner == nil && mdl.scheduledList != nil && mdl.scheduledList.Len() > 0 {
				runner = mdl.getFromSchedulledList()
				if runner.GetMovingTime() < stime {
					panic("control is seting simulation time in the past")
				} else {
					stime = runner.GetMovingTime()
				}
				mdl.addToMovingList(runner)
			}
			if runner == nil {
				break
			}
			//restarting
			mdl.activeRunner = runner
			mdl.activeRunner.setState(RUNNER_STATE_ACTIVE)
			runner.setWaitingForBoolControl(nil)
			mdl.activeRunner.getChannel() <- -1

		}
		if mdl.DEBUG {
			fmt.Println("Finished")
		}
		mdl.simulationActive = false
	}()

	return true

}
