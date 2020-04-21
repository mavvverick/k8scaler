package internal

import (
	"encoding/json"
	"fmt"
	"math"
)

type Message struct {
	Completed  int64 `json:"Completed"`
	Failed     int64 `json:"Failed"`
	Open       int64 `json:"open"`
	Timeout    int64 `json:"Timeout"`
	Cancelled  int64 `json:"Cancelled"`
	Terminated int64 `json:"Terminated"`
	Total      int64 `json:"Total"`
}

var minThreshold = int64(2)
var maxThreshold = int64(10)

//Process video data
func Process(kube KubernetesClient, msg []byte) error {
	var m Message
	if err := json.Unmarshal(msg, &m); err != nil {
		return err
	}

	deployments, err := kube.GetDeployment("cadence", "transcoder-activity")
	if err != nil {
		return err
	}

	errorJobs := m.Cancelled + m.Failed + m.Timeout + m.Terminated
	closedJobs := m.Completed + errorJobs
	totalJobs := m.Open + closedJobs
	numOfWorkers := int64(deployments.Spec.GetReplicas())
	workers := scaler(totalJobs, closedJobs, 200, numOfWorkers)
	fmt.Printf("Scale replica %v , Current %v ", workers, numOfWorkers)
	if numOfWorkers != workers {
		_, err := kube.ScaleDeployment("cadence", "transcoder-activity", int32(workers))
		if err != nil {
			return err
		}
	}

	return nil
}

func scaler(totalStartedJobs, totalClosedJobs, workerCompletionRate, numOfWorkers int64) int64 {
	if totalStartedJobs > totalClosedJobs {
		// if difference is greater than one worker capacity
		if ((totalStartedJobs-totalClosedJobs)/workerCompletionRate) > 1 &&
			(totalStartedJobs-totalClosedJobs) > numOfWorkers*workerCompletionRate {
			//
			extraWorkers := int64(math.Floor(float64((totalStartedJobs - totalClosedJobs) / workerCompletionRate)))
			if numOfWorkers+extraWorkers > maxThreshold {
				//Call Notification service too
				numOfWorkers += extraWorkers
			} else {
				numOfWorkers += extraWorkers
			}
		} else { // if the workers are being under utilized then downscale
			minWorkersRequired := totalStartedJobs / workerCompletionRate
			if minWorkersRequired < minThreshold {
				numOfWorkers = minThreshold
			} else {
				numOfWorkers = minWorkersRequired
			}
		}
	} else if totalStartedJobs <= totalClosedJobs {
		minWorkersRequired := totalStartedJobs / workerCompletionRate
		if minWorkersRequired < minThreshold {
			numOfWorkers = minThreshold
		} else {
			numOfWorkers = minWorkersRequired
		}
	}
	fmt.Println("Final number of workers", numOfWorkers)
	return numOfWorkers
}
