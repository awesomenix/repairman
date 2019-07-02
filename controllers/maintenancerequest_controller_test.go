/*
Copyright 2019 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers_test

// import (
// 	"context"
// 	//"strconv"
// 	//"strings"
// 	"fmt"
// 	"time"

// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"
// 	corev1 "k8s.io/api/core/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/types"

// 	//"k8s.io/apimachinery/pkg/util/sets"
// 	ctrl "sigs.k8s.io/controller-runtime"
// 	"sigs.k8s.io/controller-runtime/pkg/client"

// 	api "sigs.k8s.io/kubebuilder/test/project/api/v1"
// 	"sigs.k8s.io/kubebuilder/test/project/controllers"
// )

// var (
// 	// schedule is hourly on the hour
// 	onTheHour     = "0 * * * ?"
// 	errorSchedule = "obvious error schedule"
// )

// func justBeforeTheHour() time.Time {
// 	T1, err := time.Parse(time.RFC3339, "2016-05-19T09:59:00Z")
// 	if err != nil {
// 		panic("test setup error")
// 	}
// 	return T1
// }

// func topOfTheHour() time.Time {
// 	T1, err := time.Parse(time.RFC3339, "2016-05-19T10:00:00Z")
// 	if err != nil {
// 		panic("test setup error")
// 	}
// 	return T1
// }

// func justAfterTheHour() time.Time {
// 	T1, err := time.Parse(time.RFC3339, "2016-05-19T10:01:00Z")
// 	if err != nil {
// 		panic("test setup error")
// 	}
// 	return T1
// }

// func weekAfterTheHour() time.Time {
// 	T1, err := time.Parse(time.RFC3339, "2016-05-26T10:00:00Z")
// 	if err != nil {
// 		panic("test setup error")
// 	}
// 	return T1
// }

// func justBeforeThePriorHour() time.Time {
// 	T1, err := time.Parse(time.RFC3339, "2016-05-19T08:59:00Z")
// 	if err != nil {
// 		panic("test setup error")
// 	}
// 	return T1
// }

// func justAfterThePriorHour() time.Time {
// 	T1, err := time.Parse(time.RFC3339, "2016-05-19T09:01:00Z")
// 	if err != nil {
// 		panic("test setup error")
// 	}
// 	return T1
// }

// func startTimeStringToTime(startTime string) time.Time {
// 	T1, err := time.Parse(time.RFC3339, startTime)
// 	if err != nil {
// 		panic("test setup error")
// 	}
// 	return T1
// }

// // returns a cronJob with some fields filled in.
// func cronJob() api.CronJob {
// 	return api.CronJob{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      "mycronjob",
// 			Namespace: "snazzycats",
// 		},
// 		Spec: api.CronJobSpec{
// 			Schedule:          "* * * * ?",
// 			ConcurrencyPolicy: api.AllowConcurrent,
// 			JobTemplate: api.JobTemplateSpec{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Labels:      map[string]string{"a": "b"},
// 					Annotations: map[string]string{"x": "y"},
// 				},
// 				Spec: jobSpec(),
// 			},
// 		},
// 	}
// }

// func jobSpec() api.JobSpec {
// 	one := int32(1)
// 	return api.JobSpec{
// 		Parallelism: &one,
// 		Completions: &one,
// 		Template: corev1.PodTemplateSpec{
// 			ObjectMeta: metav1.ObjectMeta{
// 				Labels: map[string]string{
// 					"foo": "bar",
// 				},
// 			},
// 			Spec: corev1.PodSpec{
// 				Containers: []corev1.Container{
// 					{Image: "foo/bar"},
// 				},
// 			},
// 		},
// 	}
// }

// func newJob(UID string) api.Job {
// 	return api.Job{
// 		ObjectMeta: metav1.ObjectMeta{
// 			UID:       types.UID(UID),
// 			Name:      "foobar",
// 			Namespace: metav1.NamespaceDefault,
// 		},
// 		Spec: jobSpec(),
// 	}
// }

// var (
// 	shortDead  int64 = 10
// 	mediumDead int64 = 2 * 60 * 60
// 	longDead   int64 = 1000000
// 	noDead     int64 = -12345
// )

// type fakeClock struct {
// 	now *time.Time
// }

// func (c *fakeClock) Now() time.Time { return *c.now }

// var _ = Describe("Miaintenance Request Controller", func() {
// 	Context("when deciding when to run", func() {
// 		var cronJobController *controllers.Main
// 		var mgrStop chan struct{}
// 		var now time.Time

// 		BeforeEach(func() {
// 			mgrStop = make(chan struct{})
// 			mgr, err := ctrl.NewManager(cfg, ctrl.Options{MapperProvider: mapperProvider})
// 			Expect(err).NotTo(HaveOccurred())
// 			api.AddToScheme(mgr.GetScheme())
// 			cronJobController = &controllers.Maintean{
// 				Log:   ctrl.Log.WithName("controllers").WithName("cronjob"),
// 				Clock: &fakeClock{now: &now},
// 			}
// 			Expect(cronJobController.SetupWithManager(mgr)).To(Succeed())

// 			go func() {
// 				Expect(mgr.Start(mgrStop)).To(Succeed())
// 			}()
// 		})
// 		AfterEach(func() {
// 			close(mgrStop)
// 		})
// 		// Check expectations on deadline parameters
// 		testCases := map[string]struct {
// 			// cronJob spec
// 			concurrencyPolicy api.ConcurrencyPolicy
// 			suspend           bool
// 			schedule          string
// 			deadline          int64

// 			// cronJob status
// 			ranPreviously bool
// 			stillActive   bool

// 			// environment
// 			now time.Time

// 			// expectations
// 			expectCreate     bool
// 			expectDelete     bool
// 			expectActive     int
// 			expectedWarnings int
// 		}{
// 			/*"never ran, not valid schedule, api.AllowConcurrent":      {api.AllowConcurrent, false, errorSchedule, noDead, false, false, justBeforeTheHour(), false, false, 0, 1},
// 			"never ran, not valid schedule, false":      {api.ForbidConcurrent, false, errorSchedule, noDead, false, false, justBeforeTheHour(), false, false, 0, 1},
// 			"never ran, not valid schedule, api.ReplaceConcurrent":      {api.ForbidConcurrent, false, errorSchedule, noDead, false, false, justBeforeTheHour(), false, false, 0, 1},
// 			"never ran, not time, api.AllowConcurrent":                {api.AllowConcurrent, false, onTheHour, noDead, false, false, justBeforeTheHour(), false, false, 0, 0},
// 			"never ran, not time, false":                {api.ForbidConcurrent, false, onTheHour, noDead, false, false, justBeforeTheHour(), false, false, 0, 0},
// 			"never ran, not time, api.ReplaceConcurrent":                {api.ReplaceConcurrent, false, onTheHour, noDead, false, false, justBeforeTheHour(), false, false, 0, 0},
// 			"never ran, is time, api.AllowConcurrent":                 {api.AllowConcurrent, false, onTheHour, noDead, false, false, justAfterTheHour(), true, false, 1, 0},
// 			"never ran, is time, false":                 {api.ForbidConcurrent, false, onTheHour, noDead, false, false, justAfterTheHour(), true, false, 1, 0},
// 			"never ran, is time, api.ReplaceConcurrent":                 {api.ReplaceConcurrent, false, onTheHour, noDead, false, false, justAfterTheHour(), true, false, 1, 0},
// 			"never ran, is time, suspended":         {api.AllowConcurrent, true, onTheHour, noDead, false, false, justAfterTheHour(), false, false, 0, 0},
// 			"never ran, is time, past deadline":     {api.AllowConcurrent, false, onTheHour, shortDead, false, false, justAfterTheHour(), false, false, 0, 0},
// 			"never ran, is time, not past deadline": {api.AllowConcurrent, false, onTheHour, longDead, false, false, justAfterTheHour(), true, false, 1, 0},*/

// 			"prev ran but done, not time, api.AllowConcurrent":   {api.AllowConcurrent, false, onTheHour, noDead, true, false, justBeforeTheHour(), false, false, 0, 0},
// 			"prev ran but done, not time, api.ReplaceConcurrent": {api.ReplaceConcurrent, false, onTheHour, noDead, true, false, justBeforeTheHour(), false, false, 0, 0},
// 			"prev ran but done, is time, api.AllowConcurrent":    {api.AllowConcurrent, false, onTheHour, noDead, true, false, justAfterTheHour(), true, false, 1, 0},
// 			"prev ran but done, is time, false":                  {api.ForbidConcurrent, false, onTheHour, noDead, true, false, justAfterTheHour(), true, false, 1, 0},
// 			"prev ran but done, is time, api.ReplaceConcurrent":  {api.ReplaceConcurrent, false, onTheHour, noDead, true, false, justAfterTheHour(), true, false, 1, 0},
// 			"prev ran but done, is time, suspended":              {api.AllowConcurrent, true, onTheHour, noDead, true, false, justAfterTheHour(), false, false, 0, 0},
// 			"prev ran but done, is time, past deadline":          {api.AllowConcurrent, false, onTheHour, shortDead, true, false, justAfterTheHour(), false, false, 0, 0},
// 			"prev ran but done, is time, not past deadline":      {api.AllowConcurrent, false, onTheHour, longDead, true, false, justAfterTheHour(), true, false, 1, 0},
// 			"prev ran but done, not time, false":                 {api.ForbidConcurrent, false, onTheHour, noDead, true, false, justBeforeTheHour(), false, false, 0, 0},

// 			"still active, not time, api.AllowConcurrent":   {api.AllowConcurrent, false, onTheHour, noDead, true, true, justBeforeTheHour(), false, false, 1, 0},
// 			"still active, not time, false":                 {api.ForbidConcurrent, false, onTheHour, noDead, true, true, justBeforeTheHour(), false, false, 1, 0},
// 			"still active, not time, api.ReplaceConcurrent": {api.ReplaceConcurrent, false, onTheHour, noDead, true, true, justBeforeTheHour(), false, false, 1, 0},
// 			"still active, is time, api.AllowConcurrent":    {api.AllowConcurrent, false, onTheHour, noDead, true, true, justAfterTheHour(), true, false, 2, 0},
// 			"still active, is time, false":                  {api.ForbidConcurrent, false, onTheHour, noDead, true, true, justAfterTheHour(), false, false, 1, 0},
// 			"still active, is time, api.ReplaceConcurrent":  {api.ReplaceConcurrent, false, onTheHour, noDead, true, true, justAfterTheHour(), true, true, 1, 0},
// 			"still active, is time, suspended":              {api.AllowConcurrent, true, onTheHour, noDead, true, true, justAfterTheHour(), false, false, 1, 0},
// 			"still active, is time, past deadline":          {api.AllowConcurrent, false, onTheHour, shortDead, true, true, justAfterTheHour(), false, false, 1, 0},
// 			"still active, is time, not past deadline":      {api.AllowConcurrent, false, onTheHour, longDead, true, true, justAfterTheHour(), true, false, 2, 0},

// 			// Controller should fail to schedule these, as there are too many missed starting times
// 			// and either no deadline or a too long deadline.
// 			"prev ran but done, long overdue, not past deadline, api.AllowConcurrent":   {api.AllowConcurrent, false, onTheHour, longDead, true, false, weekAfterTheHour(), false, false, 0, 1},
// 			"prev ran but done, long overdue, not past deadline, api.ReplaceConcurrent": {api.ReplaceConcurrent, false, onTheHour, longDead, true, false, weekAfterTheHour(), false, false, 0, 1},
// 			"prev ran but done, long overdue, not past deadline, false":                 {api.ForbidConcurrent, false, onTheHour, longDead, true, false, weekAfterTheHour(), false, false, 0, 1},
// 			"prev ran but done, long overdue, no deadline, api.AllowConcurrent":         {api.AllowConcurrent, false, onTheHour, noDead, true, false, weekAfterTheHour(), false, false, 0, 1},
// 			"prev ran but done, long overdue, no deadline, api.ReplaceConcurrent":       {api.ReplaceConcurrent, false, onTheHour, noDead, true, false, weekAfterTheHour(), false, false, 0, 1},
// 			"prev ran but done, long overdue, no deadline, false":                       {api.ForbidConcurrent, false, onTheHour, noDead, true, false, weekAfterTheHour(), false, false, 0, 1},

// 			"prev ran but done, long overdue, past medium deadline, api.AllowConcurrent": {api.AllowConcurrent, false, onTheHour, mediumDead, true, false, weekAfterTheHour(), true, false, 1, 0},
// 			"prev ran but done, long overdue, past short deadline, api.AllowConcurrent":  {api.AllowConcurrent, false, onTheHour, shortDead, true, false, weekAfterTheHour(), true, false, 1, 0},

// 			"prev ran but done, long overdue, past medium deadline, api.ReplaceConcurrent": {api.ReplaceConcurrent, false, onTheHour, mediumDead, true, false, weekAfterTheHour(), true, false, 1, 0},
// 			"prev ran but done, long overdue, past short deadline, api.ReplaceConcurrent":  {api.ReplaceConcurrent, false, onTheHour, shortDead, true, false, weekAfterTheHour(), true, false, 1, 0},

// 			"prev ran but done, long overdue, past medium deadline, false": {api.ForbidConcurrent, false, onTheHour, mediumDead, true, false, weekAfterTheHour(), true, false, 1, 0},
// 			"prev ran but done, long overdue, past short deadline, false":  {api.ForbidConcurrent, false, onTheHour, shortDead, true, false, weekAfterTheHour(), true, false, 1, 0},
// 		}
// 		nsCount := 0
// 		for name := range testCases {
// 			// avoid iteration variable issues
// 			tc := testCases[name]
// 			currentNS := nsCount
// 			nsCount++
// 			It(name, func() {
// 				ctx := context.Background()
// 				now = tc.now

// 				By("creating the namespace")
// 				nsName := fmt.Sprintf("cronjob-controller-test-%v", currentNS)
// 				Expect(cl.Create(ctx, &corev1.Namespace{
// 					ObjectMeta: metav1.ObjectMeta{
// 						Name: nsName,
// 					},
// 				})).To(Succeed())
// 				defer func() {
// 					By("deleting the namespace")
// 					Expect(cl.Delete(ctx, &corev1.Namespace{
// 						ObjectMeta: metav1.ObjectMeta{
// 							Name: nsName,
// 						},
// 					})).To(Succeed())
// 				}()

// 				By("creating the cronjob")
// 				cronJob := cronJob()
// 				cronJob.Namespace = nsName
// 				cronJob.Spec.ConcurrencyPolicy = tc.concurrencyPolicy
// 				cronJob.Spec.Suspend = &tc.suspend
// 				cronJob.Spec.Schedule = tc.schedule
// 				if tc.deadline != noDead {
// 					cronJob.Spec.StartingDeadlineSeconds = &tc.deadline
// 				}

// 				var (
// 					job *api.Job
// 					err error
// 				)
// 				jobs := []api.Job{}
// 				if tc.ranPreviously {
// 					cronJob.ObjectMeta.CreationTimestamp = metav1.Time{Time: justBeforeThePriorHour()}
// 					cronJob.Status.LastScheduleTime = &metav1.Time{Time: justAfterThePriorHour()}
// 				} else {
// 					cronJob.ObjectMeta.CreationTimestamp = metav1.Time{Time: justBeforeTheHour()}
// 					Expect(tc.stillActive).To(BeFalse(), "this test case makes no sense")
// 				}
// 				Expect(cl.Create(ctx, &cronJob)).To(Succeed())

// 				if tc.ranPreviously {
// 					By("creating the existing job")
// 					job, err = cronJobController.ConstructJobForCronJob(&cronJob, cronJob.Status.LastScheduleTime.Time)
// 					Expect(err).NotTo(HaveOccurred())
// 					jobs = append(jobs, *job)
// 					Expect(cl.Create(ctx, job)).To(Succeed())
// 					if tc.stillActive {
// 						By("setting cronjob active status")
// 						Eventually(func() error {
// 							Expect(cl.Get(ctx, types.NamespacedName{Name: cronJob.Name, Namespace: cronJob.Namespace}, &cronJob)).To(Succeed())
// 							cronJob.Status.Active = []corev1.ObjectReference{{UID: job.UID}}
// 							return cl.Status().Update(ctx, &cronJob)
// 						}).Should(Succeed())
// 					} else {
// 						job.Status.Conditions = append(job.Status.Conditions, api.JobCondition{
// 							Type:   api.JobComplete,
// 							Status: corev1.ConditionTrue,
// 						})
// 						Expect(cl.Status().Update(ctx, job)).To(Succeed())
// 					}
// 				}
// 				// TODO: still around vs still active?

// 				expectedCreates := 0
// 				if tc.expectCreate {
// 					expectedCreates = 1
// 				}
// 				expectedDeletes := 0
// 				if tc.expectDelete {
// 					expectedDeletes = 1
// 				}
// 				var outJobs api.JobList
// 				Eventually(func() []api.Job {
// 					Expect(cl.List(ctx, &outJobs, client.InNamespace(cronJob.Namespace))).To(Succeed())
// 					return outJobs.Items
// 				}).Should(HaveLen(len(jobs) + expectedCreates - expectedDeletes))
// 				// TODO: this doesn't actually measure what we want

// 				for _, job := range outJobs.Items {
// 					controllerRef := metav1.GetControllerOf(&job)
// 					Expect(controllerRef).NotTo(BeNil())
// 					definitelyTrue := true
// 					Expect(controllerRef).To(Equal(&metav1.OwnerReference{
// 						APIVersion:         api.GroupVersion.String(),
// 						Kind:               "CronJob",
// 						Name:               cronJob.Name,
// 						UID:                cronJob.UID,
// 						Controller:         &definitelyTrue,
// 						BlockOwnerDeletion: &definitelyTrue,
// 					}))
// 				}

// 				Eventually(func() []corev1.ObjectReference {
// 					Expect(cl.Get(ctx, types.NamespacedName{Namespace: cronJob.Namespace, Name: cronJob.Name}, &cronJob)).To(Succeed())
// 					return cronJob.Status.Active
// 				}).Should(HaveLen(tc.expectActive))
// 			})
// 		}
// 	})
// })
