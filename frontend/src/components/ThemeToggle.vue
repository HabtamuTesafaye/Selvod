<script setup>
import { ref, onMounted } from 'vue'
import { Sun, Moon } from 'lucide-vue-next'

const isDark = ref(false)

const toggleTheme = () => {
  isDark.value = !isDark.value
  if (isDark.value) {
    document.documentElement.classList.add('dark')
    localStorage.setItem('theme', 'dark')
  } else {
    document.documentElement.classList.remove('dark')
    localStorage.setItem('theme', 'light')
  }
}

onMounted(() => {
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme === 'dark' || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  } else {
    isDark.value = false
    document.documentElement.classList.remove('dark')
  }
})
</script>

<template>
  <button 
    class="p-2 rounded-lg text-slate-500 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800 transition-colors" 
    :title="isDark ? 'Switch to Light Mode' : 'Switch to Dark Mode'"
    @click="toggleTheme"
  >
    <Sun
      v-if="isDark"
      class="w-5 h-5"
    />
    <Moon
      v-else
      class="w-5 h-5"
    />
  </button>
</template>
