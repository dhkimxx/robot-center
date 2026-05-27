package port

import "robot-center/apps/server/internal/utils"

func NormalizeMissionRobotCodes(input CreateMissionInput) []string {
	codes := make([]string, 0, len(input.RobotCodes)+1)
	if input.RobotCode != "" {
		codes = append(codes, input.RobotCode)
	}
	codes = append(codes, input.RobotCodes...)
	return utils.UniqueTrimmedStrings(codes)
}
