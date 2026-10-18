import { onUnmounted, ref, shallowRef } from 'vue'
import type { Demand, Status } from '../types'
import { stages } from '../config/catalog'

interface Options {
  getItem: (id: string) => Demand | undefined
  canStart: (item: Demand) => boolean
  onDrop: (item: Demand, status: Status) => void
}

// Pointer events support mouse, pen and touch without changing Vue's card lists.
// Touch starts on the handle so the rest of each card remains scrollable.
export function useBoardDrag(options: Options) {
  const item = shallowRef<Demand | null>(null)
  const over = ref<Status | null>(null)
  const preview = ref({ x: 0, y: 0, width: 0 })
  let pending:
    | {
        item: Demand
        id: number
        x: number
        y: number
        offsetX: number
        offsetY: number
        width: number
        board: HTMLElement
      }
    | undefined
  let point = { x: 0, y: 0 },
    frame = 0,
    suppressUntil = 0

  function updateTarget() {
    if (!pending) return
    preview.value = {
      x: point.x - pending.offsetX,
      y: point.y - pending.offsetY,
      width: pending.width,
    }
    const column = document
      .elementFromPoint(point.x, point.y)
      ?.closest<HTMLElement>('[data-board-status]')
    over.value =
      column &&
      pending.board.contains(column) &&
      stages.some((stage) => stage.value === column.dataset.boardStatus)
        ? (column.dataset.boardStatus as Status)
        : null
  }
  function edgeSpeed(value: number, start: number, end: number) {
    const edge = Math.min(48, (end - start) / 4)
    if (value < start || value > end) return 0
    return value < start + edge
      ? -14 * (1 - (value - start) / edge)
      : value > end - edge
        ? 14 * (1 - (end - value) / edge)
        : 0
  }
  function autoScroll() {
    if (!pending || !item.value) return
    const rect = pending.board.getBoundingClientRect()
    if (point.y >= rect.top && point.y <= rect.bottom)
      pending.board.scrollLeft += edgeSpeed(point.x, rect.left, rect.right)
    updateTarget()
    const column = document
      .elementFromPoint(point.x, point.y)
      ?.closest<HTMLElement>('[data-board-status]')
    const list =
      column && pending.board.contains(column)
        ? column.querySelector<HTMLElement>('.kanban-items')
        : null
    if (list) {
      const bounds = list.getBoundingClientRect()
      list.scrollTop += edgeSpeed(point.y, bounds.top, bounds.bottom)
    }
    frame = requestAnimationFrame(autoScroll)
  }
  function clear() {
    if (item.value) suppressUntil = performance.now() + 400
    cancelAnimationFrame(frame)
    document.removeEventListener('pointermove', move, true)
    document.removeEventListener('pointerup', end, true)
    document.removeEventListener('pointercancel', cancel, true)
    document.removeEventListener('keydown', keydown, true)
    window.removeEventListener('blur', cancel)
    document.body.classList.remove('kanban-dragging')
    pending = undefined
    item.value = null
    over.value = null
  }
  function cancel() {
    clear()
  }
  function keydown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      event.preventDefault()
      event.stopPropagation()
      clear()
    }
  }
  function move(event: PointerEvent) {
    if (!pending || event.pointerId !== pending.id) return
    point = { x: event.clientX, y: event.clientY }
    if (!item.value && Math.hypot(point.x - pending.x, point.y - pending.y) < 7) return
    if (!item.value) {
      item.value = pending.item
      document.body.classList.add('kanban-dragging')
      frame = requestAnimationFrame(autoScroll)
    }
    event.preventDefault()
    updateTarget()
  }
  function end(event: PointerEvent) {
    if (!pending || event.pointerId !== pending.id) return
    if (item.value) {
      event.preventDefault()
      point = { x: event.clientX, y: event.clientY }
      updateTarget()
    }
    const dragged = item.value,
      target = over.value
    clear()
    if (dragged && target && target !== dragged.status) options.onDrop(dragged, target)
  }
  function start(event: PointerEvent) {
    if (event.button !== 0 || pending || !(event.target instanceof Element)) return
    const card = event.target.closest<HTMLElement>('[data-demand-id]')
    const board = card?.closest<HTMLElement>('.kanban-board')
    const demand = card && options.getItem(card.dataset.demandId || '')
    if (!card || !board || !demand || !options.canStart(demand)) return
    if (event.pointerType !== 'mouse' && !event.target.closest('.kanban-move-handle')) return
    const rect = card.getBoundingClientRect()
    pending = {
      item: { ...demand },
      id: event.pointerId,
      x: event.clientX,
      y: event.clientY,
      offsetX: event.clientX - rect.left,
      offsetY: event.clientY - rect.top,
      width: rect.width,
      board,
    }
    document.addEventListener('pointermove', move, { capture: true, passive: false })
    document.addEventListener('pointerup', end, true)
    document.addEventListener('pointercancel', cancel, true)
    document.addEventListener('keydown', keydown, true)
    window.addEventListener('blur', cancel)
    // Release implicit touch capture so hit testing and cross-column drops work.
    if (event.target.hasPointerCapture(event.pointerId))
      event.target.releasePointerCapture(event.pointerId)
  }
  function suppressClick(event: MouseEvent) {
    if (performance.now() < suppressUntil) {
      event.preventDefault()
      event.stopPropagation()
    }
  }
  onUnmounted(clear)
  return { item, over, preview, start, cancel, suppressClick }
}
