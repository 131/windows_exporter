package hyperv

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/client"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorHypervisorVirtualProcessor Hyper-V Hypervisor Virtual Processor metrics
type collectorHypervisorVirtualProcessor struct {
	perfDataCollectorHypervisorVirtualProcessor *pdh.Collector
	perfDataObjectHypervisorVirtualProcessor    []perfDataCounterValuesHypervisorVirtualProcessor

	hypervisorVirtualProcessorTimeTotal         *prometheus.Desc
	hypervisorVirtualProcessorTotalRunTimeTotal *prometheus.Desc
	hypervisorVirtualProcessorContextSwitches   *prometheus.Desc

	dockerClient *client.Client
}

type perfDataCounterValuesHypervisorVirtualProcessor struct {
	Name string

	HypervisorVirtualProcessorGuestRunTimePercent      float64 `perfdata:"% Guest Run Time"`
	HypervisorVirtualProcessorHypervisorRunTimePercent float64 `perfdata:"% Hypervisor Run Time"`
	HypervisorVirtualProcessorTotalRunTimePercent      float64 `perfdata:"% Total Run Time"`
	HypervisorVirtualProcessorRemoteRunTimePercent     float64 `perfdata:"% Remote Run Time"`
	HypervisorVirtualProcessorCPUWaitTimePerDispatch   float64 `perfdata:"CPU Wait Time Per Dispatch"`
}

func (c *Collector) buildHypervisorVirtualProcessor() error {
	var err error

	c.perfDataCollectorHypervisorVirtualProcessor, err = pdh.NewCollector[perfDataCounterValuesHypervisorVirtualProcessor](pdh.CounterTypeRaw, "Hyper-V Hypervisor Virtual Processor", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Hypervisor Virtual Processor collector: %w", err)
	}

	c.hypervisorVirtualProcessorTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hypervisor_virtual_processor_time_total"),
		"Time that processor spent in different modes (hypervisor, guest_run, guest_idle, remote)",
		[]string{"vm", "core", "state", "container_name", "image_name"},
		nil,
	)
	c.hypervisorVirtualProcessorTotalRunTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hypervisor_virtual_processor_total_run_time_total"),
		"Time that processor spent",
		[]string{"vm", "core", "container_name", "image_name"},
		nil,
	)
	c.hypervisorVirtualProcessorContextSwitches = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hypervisor_virtual_processor_cpu_wait_time_per_dispatch_total"),
		"The average time (in nanoseconds) spent waiting for a virtual processor to be dispatched onto a logical processor.",
		[]string{"vm", "core", "container_name", "image_name"},
		nil,
	)

	c.dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to initialize Docker client: %w", err)
	}

	return nil
}

func (c *Collector) collectHypervisorVirtualProcessor(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorHypervisorVirtualProcessor.Collect(&c.perfDataObjectHypervisorVirtualProcessor)
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Hypervisor Virtual Processor metrics: %w", err)
	}

	for _, data := range c.perfDataObjectHypervisorVirtualProcessor {
		parts := strings.Split(data.Name, ":")
		if len(parts) != 2 {
			return fmt.Errorf("unexpected format of Name in Hyper-V Hypervisor Virtual Processor: %q, expected %q", data.Name, "<VM Name>:Hv VP <vcore id>")
		}

		coreParts := strings.Split(parts[1], " ")
		if len(coreParts) != 3 {
			return fmt.Errorf("unexpected format of core identifier in Hyper-V Hypervisor Virtual Processor: %q, expected %q", parts[1], "Hv VP <vcore id>")
		}

		vmName := parts[0]
		coreID := coreParts[2]

		// Retrieve Docker container information
		containerName, imageName := c.getContainerInfo(context.Background(), vmName)

		states := map[string]float64{
			"hypervisor": data.HypervisorVirtualProcessorHypervisorRunTimePercent,
			"guest":      data.HypervisorVirtualProcessorGuestRunTimePercent,
			"remote":     data.HypervisorVirtualProcessorRemoteRunTimePercent,
		}

		for state, value := range states {
			ch <- prometheus.MustNewConstMetric(
				c.hypervisorVirtualProcessorTimeTotal,
				prometheus.CounterValue,
				value,
				vmName, coreID, state, containerName, imageName,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorVirtualProcessorTotalRunTimeTotal,
			prometheus.CounterValue,
			data.HypervisorVirtualProcessorTotalRunTimePercent,
			vmName, coreID, containerName, imageName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorVirtualProcessorContextSwitches,
			prometheus.CounterValue,
			data.HypervisorVirtualProcessorCPUWaitTimePerDispatch,
			vmName, coreID, containerName, imageName,
		)
	}

	return nil
}

func (c *Collector) getContainerInfo(ctx context.Context, containerID string) (string, string) {
	containerJSON, err := c.dockerClient.ContainerInspect(ctx, containerID)
	if err != nil {
		return "unknown", "unknown"
	}

	return strings.TrimPrefix(containerJSON.Name, "/"), containerJSON.Config.Image
}
