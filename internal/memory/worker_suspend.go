package memory

func ClassifyProcess(pid uint32) string {
	return "unknown"
}

func SuspendWorkers() {
}

func ResumeWorkers() {
}

func ResumeWorkersSafe() {
	ResumeWorkers()
}
