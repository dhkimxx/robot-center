import MultiSelectField from "../../components/MultiSelectField.jsx";
import SelectField from "../../components/SelectField.jsx";
import { missionTypes } from "../../config/controlCenterConfig.js";
import { makeStatusLabel } from "../../utils/formatters.js";

export function MissionFormFields({ form, robots, setForm }) {
  return (
    <>
      <label className="grid gap-1.5 text-xs font-extrabold text-slate-400">
        <span>임무명</span>
        <input
          value={form.name}
          onChange={(event) => setForm({ ...form, name: event.target.value })}
        />
      </label>
      <SelectField
        label="시나리오"
        options={missionTypes}
        value={form.missionType}
        onChange={(missionType) => setForm({ ...form, missionType })}
      />
      <MultiSelectField
        label="배정 로봇"
        options={robots.map((robot) => ({
          value: robot.robotCode,
          label: robot.displayName || robot.robotCode,
          description: `${robot.robotCode} / ${makeStatusLabel(robot.status)}`
        }))}
        value={form.robotCodes ?? []}
        onChange={(robotCodes) => setForm({ ...form, robotCode: robotCodes[0] ?? "", robotCodes })}
      />
      <label className="grid gap-1.5 text-xs font-extrabold text-slate-400">
        <span>현장 메모</span>
        <textarea
          value={form.siteNote}
          onChange={(event) => setForm({ ...form, siteNote: event.target.value })}
        />
      </label>
    </>
  );
}
