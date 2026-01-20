package monitor

import (
	"strings"

	"github.com/shirou/gopsutil/v4/disk"
)

func filterPartitions(parts []disk.PartitionStat, filter FilterConfig) []disk.PartitionStat {
	if len(parts) == 0 {
		return parts
	}
	filtered := make([]disk.PartitionStat, 0, len(parts))
	for _, part := range parts {
		if len(filter.MountInclude) > 0 {
			if !matchMount(filter.MountInclude, part.Mountpoint) {
				continue
			}
		} else if matchMount(filter.MountExclude, part.Mountpoint) {
			continue
		}
		if containsLower(filter.FstypeExclude, part.Fstype) {
			continue
		}
		filtered = append(filtered, part)
	}
	return filtered
}

func matchMount(list []string, mountpoint string) bool {
	for _, item := range list {
		if item == "" {
			continue
		}
		if strings.HasSuffix(item, "*") {
			prefix := strings.TrimSuffix(item, "*")
			if strings.HasPrefix(mountpoint, prefix) {
				return true
			}
			continue
		}
		if mountpoint == item {
			return true
		}
	}
	return false
}

func containsLower(list []string, value string) bool {
	if value == "" {
		return false
	}
	needle := strings.ToLower(value)
	for _, item := range list {
		if item == needle {
			return true
		}
	}
	return false
}
