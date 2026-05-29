import { ref, watch } from 'vue'

export function useDebouncedRef(initialValue, delay = 300) {
  const raw = ref(initialValue)
  const debounced = ref(initialValue)
  let timer = null

  watch(raw, (val) => {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => {
      debounced.value = val
    }, delay)
  })

  return { raw, debounced }
}
