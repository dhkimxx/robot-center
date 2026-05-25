import { useId } from "react";

export default function MultiSelectField({
  disabled = false,
  id,
  label,
  onChange,
  options = [],
  value = []
}) {
  const generatedId = useId();
  const fieldId = id ?? generatedId;
  const selectedValues = new Set(value);

  function toggleOption(optionValue) {
    if (selectedValues.has(optionValue)) {
      onChange(value.filter((currentValue) => currentValue !== optionValue));
      return;
    }
    onChange([...value, optionValue]);
  }

  return (
    <fieldset className="m-0 min-w-0 border-0 p-0" id={fieldId} disabled={disabled}>
      <legend className="mb-1.5 p-0 text-xs font-extrabold text-slate-400">{label}</legend>
      <div className="grid max-h-44 gap-1.5 overflow-auto rounded-lg border border-slate-500/20 bg-command-950/30 p-1.5">
        {options.length === 0 ? (
          <p className="m-0 p-2 text-xs font-bold text-slate-500">선택 가능한 항목이 없습니다.</p>
        ) : (
          options.map((option) => (
            <label
              className="grid min-h-10 min-w-0 grid-cols-[18px_minmax(0,1fr)] items-center gap-2 rounded-lg border border-transparent bg-white/[0.045] px-2 py-1.5 text-slate-100 transition hover:border-sapphire-500/35 hover:bg-sapphire-500/[0.11]"
              key={option.value}
            >
              <input
                className="m-0 h-4 w-4 accent-sapphire-500"
                checked={selectedValues.has(option.value)}
                type="checkbox"
                value={option.value}
                onChange={() => toggleOption(option.value)}
              />
              <span className="block min-w-0">
                <strong className="block truncate text-xs font-bold text-slate-50">{option.label}</strong>
                {option.description ? <small className="block truncate text-[11px] font-bold text-slate-500">{option.description}</small> : null}
              </span>
            </label>
          ))
        )}
      </div>
    </fieldset>
  );
}
