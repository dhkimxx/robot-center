export function createInitialRobotForm() {
  return {
    displayName: "현장 로봇 1",
    modelName: "Field Robot"
  };
}

export function createRobotEditForm(robot) {
  return {
    displayName: robot?.displayName ?? "",
    modelName: robot?.modelName ?? ""
  };
}
