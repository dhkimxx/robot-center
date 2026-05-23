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
          500: "#0D71D3",
          600: "#075CB3",
          700: "#064783",
          900: "#061A33"
        },
        command: {
          950: "#050914",
          900: "#08111F",
          850: "#0B1425",
          800: "#0E1B2D",
          700: "#17263A",
          100: "#E9EEF7"
        }
      },
      fontFamily: {
        sans: ["Mulish", "Pretendard", "Noto Sans KR", "system-ui", "sans-serif"]
      },
      boxShadow: {
        command: "0 24px 80px rgba(0, 0, 0, 0.35)",
        sapphire: "0 0 0 1px rgba(13, 113, 211, 0.35), 0 18px 48px rgba(13, 113, 211, 0.16)"
      }
    }
  },
  plugins: []
};
