package metronome

import (
	"errors"
	"fmt"
	"strings"
	"time"
	duration "github.com/ChannelMeter/iso8601duration"
	"encoding/json"
	"regexp"
	"strconv"
	"github.com/Sirupsen/logrus"
	"net/http"
)

func (client *Client) CreateJob(job *Job) (*Job, error) {
	var reply Job
	if _, err := client.apiPost(MetronomeAPIJobCreate, nil, job, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

// DELETE /v1/jobs/$jobId
func (client *Client)  DeleteJob(jobId string) (interface{}, error) {
	var msg Job //json.RawMessage
	if _, err := client.apiDelete(fmt.Sprintf(MetronomeAPIJobDelete, jobId), nil, &msg); err != nil {
		return nil, err
	} else {
		return msg, err
	}
}
// GET /v1/jobs/$jobId
func (client *Client) GetJob(jobId string) (*Job, error) {
	var job Job
	queryParams := map[string][]string{
		"embed" : {
			"historySummary",
			"activeRuns",
			"schedules",
		},
	}

	if _, err := client.apiGet(fmt.Sprintf(MetronomeAPIJobGet, jobId), queryParams, &job); err != nil {
		return nil, err
	} else {
		return &job, err
	}
}
// GET /v1/jobs
func (client *Client)  Jobs() (*[]Job, error) {
	//	jobs := new(Jobs)
	jobs := make([]Job, 0, 0)
	queryParams := map[string][]string{
		"embed" : {
			"historySummary",
			"activeRuns",
		},
	}

	_, err := client.apiGet(MetronomeAPIJobList, queryParams, &jobs)

	if err != nil {
		return nil, err
	}
	return &jobs, nil
}


// PUT /v1/jobs/$jobId
func (client *Client) UpdateJob(jobId string, job *Job) (interface{}, error) {
	var msg json.RawMessage
	if _, err := client.apiPut(fmt.Sprintf(MetronomeAPIJobUpdate, jobId), nil, job, &msg); err != nil {
		if bbb, err2 := json.Marshal(msg); err2 != nil {
			return nil, errors.New(fmt.Sprintf("JobUpdate error %s\n\tAnd %s\n", err.Error(), err2.Error()))
		} else {
			return nil, errors.New(fmt.Sprintf("JobUpdate error %s\n%s\n", err, string(bbb)))
		}
	}
	return &msg, nil
}
//
// schedules
// GET /v1/jobs/$jobId/runs


func (client *Client) Runs(jobId string, since int64) (*Job, error) {
	//jobs := make([]JobStatus, 0, 0)
	//jobs := make([]Job, 0, 0)
	var jobs Job
	queryParams := map[string][]string{
		"_timestamp": {
			strconv.FormatInt(since , 10),
//			strconv.FormatInt(time.Now().UnixNano() / int64(time.Millisecond) - 24 * 3600000, 10),
		},
		"embed" : {
			"history",
			"historySummary",
			"activeRuns",
			"schedules",
		},
	}
	// lame hidden parameters are only reachable via /v1/jobs/$jobId with queryParams
	_, err := client.apiGet(fmt.Sprintf(MetronomeAPIJobGet, jobId), queryParams, &jobs)
	if err != nil {
		return nil, err
	}
	return &jobs, nil
}
// the RunLs is the way the MetronomeAPI should work.  Until it does, this is not part of the
// the Metronome interface
func (client *Client) RunLs(jobId string) (*[]JobStatus, error) {
	jobs := make([]JobStatus, 0, 0)

	_, err := client.apiGet(fmt.Sprintf(MetronomeAPIJobRunList, jobId), nil, &jobs)

	if err != nil {
		return nil, err
	}
	return &jobs, nil
}

// POST /v1/jobs/$jobId/runs
func (client *Client) StartJob(jobId string) (interface{}, error) {
	var msg JobStatus
	if _, err := client.apiPost(fmt.Sprintf(MetronomeAPIJobRunStart, jobId), nil, jobId, &msg); err != nil {
		return nil, err
	}
	return msg, nil
}
// GET /v1/jobs/$jobId/runs/$runId
func (client *Client)  StatusJob(jobId string, runId string) (*JobStatus, error) {
	var job JobStatus

	if _, err := client.apiGet(fmt.Sprintf(MetronomeAPIJobRunStatus, jobId, runId), nil, &job); err != nil {
		return nil, err
	} else {
		return &job, err
	}
}
// POST /v1/jobs/$jobId/runs/$runId/action/stop
func (client *Client) StopJob(jobId string, runId string) (interface{}, error) {
	var msg json.RawMessage
	if _, err := client.apiPost(fmt.Sprintf(MetronomeAPIJobRunStop, jobId, runId), nil, jobId, &msg); err != nil {
		return nil, err
	}
	return msg, nil
}

//
// Schedules
//
// POST /v1/jobs/$jobId/schedules
func (client *Client) CreateSchedule(jobId string, sched *Schedule) (interface{}, error) {
	var msg Schedule //json.RawMessage
	logrus.Debugf("client.JobScheduleCreate %s\n", jobId)
	if _, err := client.apiPost(fmt.Sprintf(MetronomeAPIJobScheduleCreate, jobId), nil, sched, &msg); err != nil {
		return nil, err
	}
	return msg, nil

}
// GET /v1/jobs/$jobId/schedules/$scheduleId
func (client *Client) GetSchedule(jobId string, schedId string) (*Schedule, error) {
	var sched Schedule

	if _, err := client.apiGet(fmt.Sprintf(MetronomeAPIJobScheduleStatus, jobId, schedId), nil, &sched); err != nil {
		return nil, err
	} else {
		fmt.Printf("sched: %+v\n", sched)
		return &sched, err
	}
}
// GET /v1/jobs/$jobId/schedules
func (client *Client) Schedules(jobId string) (*[]Schedule, error) {
	scheds := make([]Schedule, 0, 0)

	_, err := client.apiGet(fmt.Sprintf(MetronomeAPIJobScheduleList, jobId), nil, &scheds)

	if err != nil {
		return nil, err
	}
	return &scheds, nil
}
// DELETE /v1/jobs/$jobId/schedules/$scheduleId
func (client *Client) DeleteSchedule(jobId string, schedId string) (interface{}, error) {
	var msg json.RawMessage
	if status, err := client.apiDelete(fmt.Sprintf(MetronomeAPIJobScheduleDelete, jobId, schedId), nil, &msg); err != nil {
		return nil, err
	} else {
		if len([]byte(msg)) == 0 {
			return http.StatusText(status), nil
		}
		return msg, err
	}
}
// PUT /v1/jobs/$jobId/schedules/$scheduleId
func (client *Client) UpdateSchedule(jobId string, schedId string, sched *Schedule) (interface{}, error) {
	var msg json.RawMessage
	if _, err := client.apiPut(fmt.Sprintf(MetronomeAPIJobScheduleUpdate, jobId, schedId), nil, sched, &msg); err != nil {
		if bbb, err2 := json.Marshal(msg); err2 != nil {
			return nil, errors.New(fmt.Sprintf("JobScheduleUpdate error %s\n\tAnd %s\n", err.Error(), err2.Error()))
		} else {
			return nil, errors.New(fmt.Sprintf("JobScheduleUpdate error %s\n\tAnd%s\n", err, string(bbb)))
		}
	}
	return sched, nil

}
//  GET  /v1/metrics
func (client *Client) Metrics() (interface{}, error) {
	msg := json.RawMessage{}
	if _, err := client.apiGet(MetronomeAPIMetrics, nil, &msg); err != nil {
		return nil, err
	} else {
		return &msg, err
	}
}


//  GET /v1/ping
func (client *Client) Ping() (*string, error) {
	val := new(string)
	msg := (interface{})(val)
	if _, err := client.apiGet(MetronomeAPIPing, nil, msg); err != nil {
		return nil, err
	} else {
		// use Sprintf to reflect the value out.  painful
		retval := fmt.Sprintf("%s", *msg.(*string))
		return &retval, err
	}
}

// RunOnceNowSchedule will return a schedule that starts immediately, runs once,
// and runs every 2 minutes until successful
func RunOnceNowSchedule() string {
	return ImmediateCrontab()
}


/*

// StartJob can manually start a job
// name: The name of the job to start
// args: A map of arguments to append to the job's command
func (client *Client) StartJob(name string) error {
	raw := json.RawMessage{}
	return client.apiPost(fmt.Sprintf(MetronomeAPIStartJob, name),nil,nil,raw)

}

// AddScheduledJob will add a scheduled job
// job: The job you would like to schedule
func (client *Client) AddScheduledJob(job *Job, sched *Schedule) error {
	return client.apiPost(MetrononeAPIAddScheduledJob, nil, job, nil)
}

// RunOnceNowJob will add a scheduled job with a schedule generated by RunOnceNowSchedule
func (client *Client) RunOnceNowJob(job *Job) error {
	//job.Schedule = RunOnceNowSchedule()
	//job.Epsilon = "PT10M"
	if sched, err := ImmediateSchedule(); err != nil{
		return err
	} else {
		return client.AddScheduledJob(job, sched)
	}
}
*/

func formatTimeString(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.Format(time.RFC3339Nano)
}


///
var (
	repeatRegex = regexp.MustCompile(`R((?P<repeat>\d*))`)
)

func ConvertIso8601ToCron(isoRep string) (string, error) {
	pat := strings.Split(isoRep, "/")
	if len(pat) == 3 {
		interval := pat[2]
		dur := pat[0]
		repeatTimes := 0
		if repeatRegex.MatchString(dur) {
			match := repeatRegex.FindStringSubmatch(dur)
			for i, name := range repeatRegex.SubexpNames() {
				part := match[i]
				if i == 0 || name == "" || part == "" {
					continue
				}
				val, err := strconv.Atoi(part)
				if err != nil {
					return "", err
				}
				switch name {
				case "repeat":
					repeatTimes = val
				default:
					return "", errors.New(fmt.Sprintf("unknown field %s", name))
				}
			}
		} else {
			return "", errors.New(fmt.Sprintf("No repeat pattern"))

		}
		tdur, err := duration.FromString(interval)

		if err != nil {
			return "", errors.New("Illegal duration")
		}
		time_t := tdur.ToDuration()
		if repeatTimes != 0 {
			// minute is the smallest scheduling unit for metronome
			slot := int64(time_t)
			if slot < 1 {
				return "", errors.New("Too small a duration")
			} else if slot < 60 {

			}

		} else {

		}

	} else {
		var (
			y, M, d, h, m, s int
		)
		if _, err := fmt.Sscanf(time.Now().Format(time.RFC3339), "%d-%d-%dT%d:%d:%dZ", &y, &M, &d, &h, &m, &s); err != nil {
			return "", err
		} else {
			return fmt.Sprint("%d %d %d %d * %d%", m, h, d, M, y), nil
		}
	}

	return "", errors.New("Unknown error")
}

func ImmediateCrontab() string {
	var (
		y, M, d, h, m, s int
	)
	if _, err := fmt.Sscanf(time.Now().Format(time.RFC3339), "%d-%d-%dT%d:%d:%dZ", &y, &M, &d, &h, &m, &s); err != nil {
		return ""
	} else {
		return fmt.Sprint("%d %d %d %d * %d%", m, h, d, M, y)
	}

}
func ImmediateSchedule() (*Schedule, error) {
	var (
		y, M, d, h, m, s int
		cronstr string
	)
	if _, err := fmt.Sscanf(time.Now().Format(time.RFC3339), "%d-%d-%dT%d:%d:%dZ", &y, &M, &d, &h, &m, &s); err != nil {
		return nil, err
	} else {
		cronstr = fmt.Sprint("%d %d %d %d * %d%", m, h, d, M, y)
	}

	sched := &Schedule{
		ID:  fmt.Sprintf("%s-%d%d%d%d%d%d", y, M, d, h, m, s), //"everyminute",
		Cron: cronstr, //"cron": "* * * * *",
		ConcurrencyPolicy: "ALLOW",
		Enabled: true,
		StartingDeadlineSeconds:60,
		Timezone: "GMT",
	}
	return sched, nil
}
