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
    <fieldset className="multi-select-field" id={fieldId} disabled={disabled}>
      <legend>{label}</legend>
      <div className="multi-select-options">
        {options.length === 0 ? (
          <p className="multi-select-empty">선택 가능한 항목이 없습니다.</p>
        ) : (
          options.map((option) => (
            <label className="multi-select-option" key={option.value}>
              <input
                checked={selectedValues.has(option.value)}
                type="checkbox"
                value={option.value}
                onChange={() => toggleOption(option.value)}
              />
              <span>
                <strong>{option.label}</strong>
                {option.description ? <small>{option.description}</small> : null}
              </span>
            </label>
          ))
        )}
      </div>
    </fieldset>
  );
}
