/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,jsx,ts,tsx}"],
  theme: {
    extend: {
      colors: {
        sapphire: {
          50: "#EAF4FF",
          100: "#D4E9FF",
          300: "#72B7F3",
          500: "#3B82F6",
          600: "#2563EB",
          700: "#1D4ED8",
          900: "#1E3A8A"
        },
        command: {
          950: "#090B0F",
          900: "#0F1218",
          850: "#121720",
          800: "#151A22",
          700: "#1D2430",
          100: "#F4F7FB"
        }
      },
      fontFamily: {
        sans: ["Mulish", "Pretendard", "Noto Sans KR", "system-ui", "sans-serif"]
      },
      boxShadow: {
        command: "0 16px 36px rgba(0, 0, 0, 0.22)",
        sapphire: "0 0 0 1px rgba(59, 130, 246, 0.18)"
      }
    }
  },
  plugins: []
};
