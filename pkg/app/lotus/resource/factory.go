package resource

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	lotusv1beta1 "github.com/nghialv/lotus/pkg/app/lotus/apis/lotus/v1beta1"
	"github.com/nghialv/lotus/pkg/app/lotus/config"
	"github.com/nghialv/lotus/pkg/app/lotus/model"
	"github.com/nghialv/lotus/pkg/version"
)

var (
	thanosImage     = "improbable/thanos:v0.2.0"
	prometheusImage = "quay.io/prometheus/prometheus:v2.3.2"
	lotusImage      = fmt.Sprintf("nghialv2607/lotus:%s", version.Get().GitCommit)
)

type ResourceFactory interface {
	PreparerJobName() string
	MonitorJobName() string
	CleanerJobName() string
	WorkerName() string
	PrometheusName() string

	NewPreparerJob() (*batchv1.Job, error)
	NewCleanerJob() (*batchv1.Job, error)
	NewMonitorJob() (*batchv1.Job, error)
	NewMonitorConfigMap() (*corev1.ConfigMap, error)
	NewWorkerDeployment() (*appsv1.Deployment, error)
	NewWorkerService() (*corev1.Service, error)
	NewPrometheusPod(serviceAccountName, release string) (*corev1.Pod, error)
	NewPrometheusService() (*corev1.Service, error)
	NewPrometheusConfigMap() (*corev1.ConfigMap, error)
}

type resourceFactory struct {
	lotus      *lotusv1beta1.Lotus
	configFile string
}

func NewFactory(lotus *lotusv1beta1.Lotus, configFile string) ResourceFactory {
	return &resourceFactory{
		lotus:      lotus,
		configFile: configFile,
	}
}

func (rf *resourceFactory) PreparerJobName() string {
	return jobName(rf.lotus.Name, JobPreparer)
}

func (rf *resourceFactory) MonitorJobName() string {
	return jobName(rf.lotus.Name, JobMonitor)
}

func (rf *resourceFactory) CleanerJobName() string {
	return jobName(rf.lotus.Name, JobCleaner)
}

func (rf *resourceFactory) WorkerName() string {
	return workerName(rf.lotus.Name)
}

func (rf *resourceFactory) PrometheusName() string {
	return prometheusName(rf.lotus.Name)
}

func (rf *resourceFactory) NewPreparerJob() (*batchv1.Job, error) {
	return newJob(
		rf.lotus,
		rf.lotus.Spec.Preparer.Containers,
		rf.lotus.Spec.Preparer.Volumes,
		JobPreparer,
	), nil
}

func (rf *resourceFactory) NewCleanerJob() (*batchv1.Job, error) {
	return newJob(
		rf.lotus,
		rf.lotus.Spec.Cleaner.Containers,
		rf.lotus.Spec.Cleaner.Volumes,
		JobCleaner,
	), nil
}

func (rf *resourceFactory) NewMonitorJob() (*batchv1.Job, error) {
	cfg, err := config.FromFile(rf.configFile)
	if err != nil {
		return nil, err
	}
	return newMonitorJob(rf.lotus, cfg), nil
}

func (rf *resourceFactory) NewMonitorConfigMap() (*corev1.ConfigMap, error) {
	cfg, err := buildLotusConfig(rf.configFile, rf.lotus)
	if err != nil {
		return nil, err
	}
	data, err := cfg.MarshalToYaml()
	if err != nil {
		return nil, err
	}
	return newMonitorConfigMap(rf.lotus, data), nil
}

func (rf *resourceFactory) NewWorkerDeployment() (*appsv1.Deployment, error) {
	return newWorkerDeployment(rf.lotus), nil
}

func (rf *resourceFactory) NewWorkerService() (*corev1.Service, error) {
	return newWorkerService(rf.lotus), nil
}

func (rf *resourceFactory) NewPrometheusPod(serviceAccountName, release string) (*corev1.Pod, error) {
	cfg, err := buildLotusConfig(rf.configFile, rf.lotus)
	if err != nil {
		return nil, err
	}
	return newPrometheusPod(rf.lotus, serviceAccountName, release, cfg)
}

func (rf *resourceFactory) NewPrometheusService() (*corev1.Service, error) {
	return newPrometheusService(rf.lotus), nil
}

func (rf *resourceFactory) NewPrometheusConfigMap() (*corev1.ConfigMap, error) {
	cfg, err := config.FromFile(rf.configFile)
	if err != nil {
		return nil, err
	}
	target := workerName(rf.lotus.Name)
	return newPrometheusConfigMap(rf.lotus, target, cfg.LotusChecks())
}

func buildLotusConfig(configFile string, lotus *lotusv1beta1.Lotus) (*config.Config, error) {
	cfg, err := config.FromFile(configFile)
	if err != nil {
		return nil, err
	}
	cfg.DataSources = append(cfg.DataSources, clientPrometheusDataSource(lotus))
	cfg.AddChecks(lotus.Spec.Checks...)
	for i := range cfg.Checks {
		if cfg.Checks[i].DataSource == "" {
			cfg.Checks[i].DataSource = localPrometheusDataSourceName
		}
	}
	return cfg, nil
}

func ownerReferences(lotus *lotusv1beta1.Lotus) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(lotus, model.ControllerKind),
	}
}
