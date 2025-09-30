package metrics_export

import (
	"context"
	"time"

	"github.com/containers/podman/v5/pkg/bindings/containers"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"

	cm "volpe-framework/container_mgr"
)

type PodmanMetricExporter struct {
	conn       context.Context
	problems   map[string]string
	deviceName string
	meter      otelmetric.Meter
}

var meterprovider *metric.MeterProvider
var exporter *otlpmetricgrpc.Exporter
var reader *metric.PeriodicReader

func InitOTelSDK() error {
	var err error
	exporter, err = otlpmetricgrpc.New(
		context.Background(),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		log.Error().Caller().Msg(err.Error())
		return err
	}
	reader = metric.NewPeriodicReader(exporter, metric.WithInterval(3*time.Second))
	meterprovider = metric.NewMeterProvider(metric.WithReader(reader))

	otel.SetMeterProvider(meterprovider)
	return nil
}

func NewPodmanMetricExporter(deviceName string, problems map[string]string) (pme *PodmanMetricExporter, err error) {
	// problems: map of problem name to container name
	pme = new(PodmanMetricExporter)
	pme.conn, err = cm.NewPodmanConnection()
	if err != nil {
		log.Error().Caller().Msg(err.Error())
		return nil, err
	}

	pme.problems = make(map[string]string)
	for problem, contName := range problems {
		pme.problems[contName] = problem
	}

	pme.deviceName = deviceName
	pme.meter = otel.Meter("volpe-framework")
	return
}

func (pme *PodmanMetricExporter) Run() error {
	conn, err := cm.NewPodmanConnection()
	if err != nil {
		log.Error().Caller().Msg(err.Error())
		return err
	}

	cpuUtilPerAppln, _ := pme.meter.Float64Gauge("volpe_cpu_util_per_appln",
		otelmetric.WithDescription("CPU Utilization per appln per container"),
	)

	statsStream := true
	statsAll := false
	contNames := make([]string, len(pme.problems))
	i := 0
	for k := range pme.problems {
		contNames[i] = k
		i += 1
	}
	statChan, _ := containers.Stats(conn, contNames, &containers.StatsOptions{
		Stream: &statsStream,
		All:    &statsAll,
	})
	for {
		report := <-statChan
		log.Info().Caller().Msg("got report")
		for i := range report.Stats {
			contName := report.Stats[i].Name
			log.Info().Caller().Msgf("reporting on %s", contName)
			pblmName, ok := pme.problems[contName]
			if !ok {
				pblmName = "<unknown>"
				log.Warn().Caller().Msgf("unknown container %s", contName)
			}
			attribSet := attribute.NewSet(
				attribute.KeyValue{Key: attribute.Key("host"), Value: attribute.StringValue(pme.deviceName)},
				attribute.KeyValue{Key: attribute.Key("problem"), Value: attribute.StringValue(pblmName)},
			)
			cpuUtilPerAppln.Record(context.Background(), report.Stats[i].CPU,
				otelmetric.WithAttributeSet(attribSet),
			)
		}
	}
}
