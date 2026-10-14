// useSelect lets the visitor pick a drawing on the chart and change it, like
// TradingView: click a drawing to select it, drag it to move it, drag one of
// its points to reshape it, and Delete removes it. Locked drawings stay put.
import { ref } from 'vue'
import { hitDrawing } from './hitTest.js'

// layer is the DrawingLayer; get.chart() and get.series() are the chart and
// the series on screen; canEdit() is false while a tool is picked or the
// drawings are hidden; changed() is called after a drawing is removed.
export function useSelect(layer, get, canEdit, changed) {
  const selected = ref(null) // index of the selected drawing
  let drag = null // { index, part, start, orig } while a drawing is dragged
  let onMove = null
  const onUp = () => (drag = null)

  // select marks a drawing as selected (null: none).
  function select(i) {
    selected.value = i
    layer.selected = i
    layer.update()
  }

  // pointAt turns a position on the chart into { time, price }.
  function pointAt(x, y) {
    const logical = get.chart().timeScale().coordinateToLogical(x)
    const price = get.series()?.coordinateToPrice(y)
    if (logical == null || price == null) return null
    return { time: layer.t0 + logical * layer.step, price }
  }

  // moved is the drawing as it was when the drag started, shifted by dt (time)
  // and dp (price): the whole drawing, one of its points, or a channel's width.
  function moved(orig, part, dt, dp) {
    const shift = (p) => p && { time: p.time + dt, price: p.price + dp }
    if (part === 'offset') return { ...orig, offset: (orig.offset || 0) + dp }
    if (part === 'p1' || part === 'p2') return { ...orig, [part]: shift(orig[part]) }
    return { ...orig, p1: shift(orig.p1), p2: shift(orig.p2) }
  }

  // remove deletes the selected drawing.
  function remove() {
    if (selected.value == null) return
    layer.drawings.splice(selected.value, 1)
    select(null)
    changed()
  }

  // install listens on the chart element before the chart does (capture), so
  // a press on a drawing drags the drawing instead of scrolling the chart.
  function install(el) {
    const local = (e) => {
      const box = el.getBoundingClientRect()
      return { x: e.clientX - box.left, y: e.clientY - box.top }
    }
    el.addEventListener('pointerdown', (e) => {
      if (!canEdit() || e.button !== 0) return
      const { x, y } = local(e)
      const hit = hitDrawing(layer, x, y, layer.size)
      if (!hit) return selected.value != null && select(null)
      select(hit.index)
      drag = { ...hit, start: pointAt(x, y), orig: { ...layer.drawings[hit.index] } }
      e.stopPropagation()
    }, true)
    // The chart scrolls on mouse and touch presses; keep those away from it
    // while a drawing is being dragged.
    for (const type of ['mousedown', 'touchstart']) {
      el.addEventListener(type, (e) => drag && e.stopPropagation(), true)
    }
    onMove = (e) => {
      if (!drag || !drag.start) return
      const { x, y } = local(e)
      const p = pointAt(x, y)
      if (!p) return
      layer.drawings[drag.index] = moved(drag.orig, drag.part, p.time - drag.start.time, p.price - drag.start.price)
      layer.update()
    }
    window.addEventListener('pointermove', onMove)
    window.addEventListener('pointerup', onUp)
  }

  // uninstall stops listening on the window (when the chart is removed).
  function uninstall() {
    window.removeEventListener('pointermove', onMove)
    window.removeEventListener('pointerup', onUp)
  }

  return { selected, select, remove, install, uninstall }
}
