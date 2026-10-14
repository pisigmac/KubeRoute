import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./src/**/*.{js,ts,jsx,tsx,mdx}"],
  theme: {
    extend: {
      fontFamily: {
        sans: ["Inter", "system-ui", "sans-serif"],
      },
      colors: {
        background: "#F5F0EB",
        foreground: "#1A1A1A",
        primary: "#1A1A1A",
        accent: "#00A854",
      },
    },
  },
  plugins: [],
};
export default config;
