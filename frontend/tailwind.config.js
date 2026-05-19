/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: 'var(--primary-color, #3B82F6)',
        accent: 'var(--accent-color, #F59E0B)',
      }
    },
  },
  plugins: [],
}
