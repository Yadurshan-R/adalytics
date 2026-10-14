// useDrawings connects the drawing toolbar to the chart: it turns clicks into
// drawings, follows the mouse while a drawing is placed, snaps to prices when
// the magnet is on, and zooms. Picking and moving a drawing is in useSelect.
// Nothing is saved: drawings live in memory.
import { ref, computed, watch } from 'vue'
import { DrawingLayer } from './drawingLayer.js'
import { useSelect } from './useSelect.js'

// get.chart() is the chart, get.series() the series on screen (line or
// candles), get.bars() the bars as [{ time, values: [prices] }].
export function useDrawings(get) {
  const layer = new DrawingLayer()
  const tool = ref('cursor') // cursor, trend, ray, extended, hline, vline, channel, rect, fib, text, ruler
  const magnet = ref(false)
  const hidden = ref(false)
  const locked = ref(false) // drawings cannot be picked or moved
  const count = ref(0) // drawings on the chart, the ruler included
  const textBox = ref(null) // { x, y, point, value } while text is typed
  const placing = ref(false) // first point clicked, second not yet
  const busy = computed(() => tool.value !== 'cursor' || placing.value || !!textBox.value)
  let series = null
  // editable is true when a click may pick a drawing: no tool picked, and the
  // drawings are shown and not locked.
  const editable = () => tool.value === 'cursor' && !hidden.value && !locked.value
  const sel = useSelect(layer, get, editable, () => recount())

  // attach puts the layer on the series on screen; call after switching
  // between line and candles.
  function attach() {
    const s = get.series()
    if (!s || s === series) return
    series?.detachPrimitive(layer)
    s.attachPrimitive(layer)
    series = s
  }

  // setBars tells the layer the time of the first bar and the bar spacing.
  function setBars(times) {
    if (!times.length) return
    layer.setBars(times[0], times.length > 1 ? times[1] - times[0] : 1)
    layer.update()
  }

  // pointFrom turns a mouse position into { time, price }. With the magnet on,
  // the price snaps to the nearest value of the bar under the mouse.
  function pointFrom(param) {
    if (!series || !param.point || param.logical == null) return null
    const index = Math.round(param.logical)
    let price = series.coordinateToPrice(param.point.y)
    if (price == null) return null
    const bars = get.bars()
    const bar = bars[Math.max(0, Math.min(bars.length - 1, index))]
    if (magnet.value && bar) {
      let best = Infinity
      for (const v of bar.values) {
        const y = series.priceToCoordinate(v)
        if (y != null && Math.abs(y - param.point.y) < best) {
          best = Math.abs(y - param.point.y)
          price = v
        }
      }
      return { time: bar.time, price }
    }
    return { time: layer.t0 + index * layer.step, price }
  }

  // onClick places drawings: one click for a horizontal or vertical line or
  // text, three for a channel, two for the others. In cursor mode a click
  // clears the ruler.
  function onClick(param) {
    const p = pointFrom(param)
    if (!p) return
    const t = tool.value
    if (t === 'cursor') {
      if (layer.ruler) clearRuler()
      return
    }
    if (t === 'hline' || t === 'vline') return finish({ type: t, p1: p })
    if (t === 'text') {
      textBox.value = { x: param.point.x, y: param.point.y, point: p, value: '' }
      return
    }
    if (!layer.draft) {
      layer.draft = { type: t, p1: p, p2: p }
      placing.value = true
      return layer.update()
    }
    // A channel's second click fixes the first line; the third sets its width.
    if (t === 'channel' && !layer.draft.stage) {
      layer.draft = { ...layer.draft, p2: p, stage: 2, offset: 0 }
      return layer.update()
    }
    const d = t === 'channel' ? { ...layer.draft } : { ...layer.draft, p2: p }
    delete d.stage
    layer.draft = null
    placing.value = false
    if (t === 'ruler') {
      layer.ruler = d
      tool.value = 'cursor'
      recount()
      return layer.update()
    }
    finish(d)
  }

  // onMove moves the second point of the drawing being placed, or for a
  // channel's last step, how far the second line is from the first.
  function onMove(param) {
    const d = layer.draft
    if (!d) return
    const p = pointFrom(param)
    if (!p) return
    if (d.stage === 2) d.offset = p.price - priceOn(d, p.time)
    else d.p2 = p
    layer.update()
  }

  // finish adds a drawing and goes back to the cursor, like TradingView.
  function finish(d) {
    layer.drawings.push(d)
    recount()
    tool.value = 'cursor'
    layer.update()
  }

  // submitText adds the typed text where it was clicked.
  function submitText() {
    const box = textBox.value
    textBox.value = null
    if (box && box.value.trim()) finish({ type: 'text', p1: box.point, text: box.value.trim().slice(0, 80) })
    else tool.value = 'cursor'
  }

  // cancel stops whatever is being drawn (the Esc key).
  function cancel() {
    layer.draft = null
    placing.value = false
    textBox.value = null
    tool.value = 'cursor'
    layer.update()
  }

  function clearRuler() {
    layer.ruler = null
    recount()
    layer.update()
  }

  const recount = () => (count.value = layer.drawings.length + (layer.ruler ? 1 : 0))

  // clearAll removes every drawing and the ruler.
  function clearAll() {
    layer.drawings = []
    sel.select(null)
    clearRuler()
    cancel()
  }

  // zoom shows fewer bars (factor below 1) or more (above 1), keeping the
  // newest bar on the right.
  function zoom(factor) {
    const ts = get.chart().timeScale()
    const r = ts.getVisibleLogicalRange()
    if (!r) return
    const width = Math.max(10, (r.to - r.from) * factor)
    ts.setVisibleLogicalRange({ from: r.to - width, to: r.to })
  }

  const resetView = () => get.chart().timeScale().fitContent()

  // priceOn is the price on the line from d.p1 to d.p2 at a given time.
  function priceOn(d, time) {
    const span = d.p2.time - d.p1.time
    return span ? d.p1.price + ((d.p2.price - d.p1.price) * (time - d.p1.time)) / span : d.p1.price
  }

  // While a tool is picked, dragging must not move the chart, so clicks place points.
  watch(tool, (t) => {
    if (t !== layer.draft?.type) {
      layer.draft = null
      placing.value = false
    }
    if (t === 'ruler') clearRuler()
    if (t !== 'cursor') hidden.value = false // a new drawing shows everything again
    get.chart().applyOptions({ handleScroll: t === 'cursor' ? SCROLL : false })
    layer.update()
  })
  watch(hidden, (h) => {
    layer.hidden = h
    layer.update()
  })
  // Nothing can be picked while a tool is on or the drawings are hidden or locked.
  watch([tool, hidden, locked], () => {
    layer.movable = editable()
    if (!layer.movable) sel.select(null)
  })

  // install listens to clicks on the chart element directly: the chart's own
  // click event drops a second click made within half a second of the first,
  // which happens all the time when placing a two-point drawing. A press that
  // moved more than a few pixels was a drag (scrolling the chart), not a click.
  function install(chart, el) {
    sel.install(el)
    let down = null
    el.addEventListener('pointerdown', (e) => (down = { x: e.clientX, y: e.clientY }))
    el.addEventListener('pointerup', (e) => {
      if (!down || Math.abs(e.clientX - down.x) + Math.abs(e.clientY - down.y) > 6) return
      const box = el.getBoundingClientRect()
      const x = e.clientX - box.left
      const y = e.clientY - box.top
      if (x > chart.timeScale().width() || y > chart.paneSize().height) return // on an axis
      onClick({ point: { x, y }, logical: chart.timeScale().coordinateToLogical(x) })
    })
    chart.subscribeCrosshairMove(onMove)
  }

  return {
    tool, magnet, hidden, locked, count, textBox, busy, selected: sel.selected,
    install, uninstall: sel.uninstall, removeSelected: sel.remove, deselect: () => sel.select(null), attach, setBars, submitText, cancel, clearAll, zoom, resetView,
  }
}

// SCROLL lets the chart be dragged sideways (but not scrolled with the mouse
// wheel, which stays for scrolling the page).
export const SCROLL = { mouseWheel: false, pressedMouseMove: true, horzTouchDrag: true, vertTouchDrag: false }
// SCALE lets the chart zoom with the mouse wheel, a pinch or by dragging the axes.
export const SCALE = { mouseWheel: true, pinch: true, axisPressedMouseMove: true, axisDoubleClickReset: true }
