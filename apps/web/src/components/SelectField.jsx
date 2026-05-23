import { useId } from "react";

export default function SelectField({
  disabled = false,
  id,
  label,
  onChange,
  options = [],
  placeholder,
  value
}) {
  const generatedId = useId();
  const selectId = id ?? generatedId;

  return (
    <label className="select-field" htmlFor={selectId}>
      <span className="select-field-label">{label}</span>
      <select
        disabled={disabled}
        id={selectId}
        value={value}
        onChange={(event) => onChange(event.target.value)}
      >
        {placeholder ? <option value="">{placeholder}</option> : null}
        {options.map((option) => (
          <option value={option.value} key={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </label>
  );
}
