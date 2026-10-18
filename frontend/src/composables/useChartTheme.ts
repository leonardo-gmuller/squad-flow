import { onMounted, onUnmounted, shallowRef } from 'vue'

export function useChartTheme() {
  const colors = shallowRef({
    text: '#8993a8',
    grid: '#e7eaf1',
    primary: '#6366f1',
    track: '#ececf5',
  })
  let observer: MutationObserver | undefined

  onMounted(() => {
    const root = document.documentElement
    function update() {
      const style = getComputedStyle(root)
      colors.value = {
        text: style.getPropertyValue('--sf-muted').trim() || '#8993a8',
        grid: style.getPropertyValue('--sf-border').trim() || '#e7eaf1',
        primary: style.getPropertyValue('--p-primary-color').trim() || '#6366f1',
        track: style.getPropertyValue('--sf-border').trim() || '#ececf5',
      }
    }
    update()
    observer = new MutationObserver(update)
    observer.observe(root, { attributes: true, attributeFilter: ['class'] })
  })

  onUnmounted(() => observer?.disconnect())
  return colors
}
