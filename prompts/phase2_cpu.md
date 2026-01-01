# Phase 2: CPU Governor (Adaptive QoS)

Implement safe CPU management with EcoQoS.

## Prompt

```
Implement 'Adaptive QoS' in internal/cpu.

Requirements:
- Define SetProcessInformation with ProcessPowerThrottling
- EnforceEfficiencyMode: EcoQoS ON + IDLE_PRIORITY_CLASS
- EnforcePerformanceMode: EcoQoS OFF + NORMAL_PRIORITY_CLASS
- Edge-triggered state machine (only apply on state change)
- Safety: Skip Performance Mode if RAM > 2GB
```

## Key APIs
- `SetProcessInformation` - EcoQoS control
- `SetPriorityClass` - Process priority
- `GetProcessMemoryInfo` - RAM check
