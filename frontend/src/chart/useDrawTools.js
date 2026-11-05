// useDrawTools connects the toolbar to lightweight-charts-drawing, the
// library that draws, places, picks and moves the drawings (TradingView's
// tools, with TradingView's mouse and keys: Esc, Delete, Shift for 45°).
// This file keeps the toolbar's state and does the few things the library
// leaves to the app: lock all, hide all, the Measure tool, the text box, and
// keeping the drawings when the chart switches between candles and line.
// Nothing is saved: drawings live in memory and go when the page closes.
import { ref } from 'vue'
import { DrawingManager } from 'lightweight-charts-drawing'

// MEASURE is the toolbar's Measure tool: a date and price range that goes
// away on the next click, like TradingView's ruler.
export const MEASURE = 'measure'
const RANGE = 'date-and-price-range'
const FONT = 'Inter, system-ui, sans-serif'

// newManager builds the library's DrawingManager; tests pass a fake instead.
const newManager = (chart, series, opts) => new DrawingManager(chart, series, opts)

// get.chart() is the chart, get.series() the series on screen (candles or
// line), get.bars() the candles [{ time, open, high, low, close, volume }].
export function useDrawTools(get, make = newManager) {
  const tool = ref('cursor') // 'cursor', MEASURE or a drawing kind (toolList.js)
  const magnet = ref('off') // 'off', 'weak' or 'strong'
  const stay = ref(false) // keep the tool after a drawing, to draw several in a row
  const locked = ref(false) // every drawing locked: no moving, no Delete key
  const hidden = ref(false) // every drawing hidden
  const count = ref(0) // drawings on the chart, the measurement not counted
  const selection = ref([]) // ids of the selected drawings
  const textBox = ref(null) // { id, x, y, value, fresh } while a text is typed
  let dm = null // the library's manager, on the series on screen
  let series = null
  let offs = [] // unsubscribe functions of the manager's events
  let measuring = false // the Measure tool is placing its range
  let measureId = null // the measurement on the chart, if any

  // attach puts the drawings on the series on screen. After a switch between
  // candles and line it builds a manager on the new series and moves every
  // drawing over (through the library's JSON, in memory only).
  function attach() {
    const s = get.series()
    if (!s || s === series) return
    const kept = dm ? dm.exportJSON() : null
    detach()
    series = s
    dm = make(get.chart(), s, { magnet: magnet.value, stayInDrawingMode: stay.value, bars: get.bars, fontFamily: FONT })
    if (kept) dm.importJSON(kept)
    dm.setDrawingsVisible(!hidden.value)
    offs = [
      dm.on('change', recount),
      dm.on('add', onAdd),
      dm.on('selection', (ids) => (selection.value = [...ids])),
      dm.on('tool', (kind) => kind || done()),
      dm.on('visibility', (shown) => (hidden.value = !shown)),
      dm.on('textEdit', onTextEdit),
    ]
    recount()
  }

  // detach removes the manager (its drawings go with it).
  function detach() {
    offs.forEach((off) => off())
    offs = []
    dm?.destroy()
    dm = null
    series = null
  }

  // recount counts the drawings, without the measurement.
  function recount() {
    if (measureId && !dm.get(measureId)) measureId = null
    count.value = dm.drawings().filter((d) => d.id !== measureId).length
  }

  // onAdd: a measurement ends the Measure tool; while "lock all" is on, a new
  // drawing is locked too.
  function onAdd(d) {
    if (measuring) {
      measureId = d.id
      measuring = false
      dm.setTool(null)
    } else if (locked.value) dm.update({ ...d, locked: true })
    recount()
  }

  // done: the library put the tool down (a drawing was placed, or Esc).
  function done() {
    measuring = false
    tool.value = 'cursor'
  }

  // pick chooses a tool. Picking the tool that is on goes back to the cursor
  // when toggle is true (a toolbar button), not from a menu.
  function pick(kind, toggle = true) {
    if (!dm) return
    clearMeasure()
    if (kind === 'cursor' || (toggle && kind === tool.value)) {
      dm.setTool(null)
      return done()
    }
    measuring = kind === MEASURE
    tool.value = kind
    dm.setTool(measuring ? RANGE : kind)
  }

  // clearMeasure removes the measurement, if there is one.
  function clearMeasure() {
    if (measureId && dm?.get(measureId)) dm.remove(measureId)
    measureId = null
  }

  // setMagnet: 'off', 'weak' (snaps when near a price) or 'strong' (always snaps).
  function setMagnet(mode) {
    magnet.value = mode
    dm?.setMagnet(mode)
  }

  function setStay(on) {
    stay.value = on
    dm?.setStayInDrawingMode(on)
  }

  // setLocked locks or unlocks every drawing (they can still be selected).
  function setLocked(on) {
    locked.value = on
    if (!dm) return
    for (const d of dm.drawings()) if (d.id !== measureId) dm.update({ ...d, locked: on })
  }

  function setHidden(on) {
    hidden.value = on
    dm?.setDrawingsVisible(!on)
  }

  // removeAll removes every drawing, locked ones included.
  function removeAll() {
    textBox.value = null
    measureId = null
    dm?.clear()
  }

  // removeSelected removes the selected drawings that are not locked.
  function removeSelected() {
    for (const id of selection.value) if (!dm.get(id)?.locked) dm.remove(id)
  }

  // onTextEdit opens the text box where a text was placed or double-clicked.
  function onTextEdit(d, at) {
    textBox.value = { id: d.id, x: at.x, y: at.y, value: d.text ?? '', fresh: !d.text }
  }

  // submitText writes the typed text on the drawing; an empty new text is removed.
  function submitText(typed) {
    const box = textBox.value
    textBox.value = null
    const d = box && dm?.get(box.id)
    if (!d) return
    const text = String(typed ?? '').trim().slice(0, 80)
    if (text) dm.update({ ...d, text })
    else if (box.fresh) dm.remove(d.id)
  }

  // cancelText closes the text box (Esc); a new text is removed.
  function cancelText() {
    const box = textBox.value
    textBox.value = null
    if (box?.fresh && dm?.get(box.id)) dm.remove(box.id)
  }

  // A click on the chart removes the measurement, like TradingView's ruler.
  const onPress = () => !measuring && clearMeasure()
  let el = null
  function install(element) {
    el = element
    el.addEventListener('pointerdown', onPress)
  }
  function uninstall() {
    el?.removeEventListener('pointerdown', onPress)
    detach()
  }

  return {
    tool, magnet, stay, locked, hidden, count, selection, textBox,
    attach, install, uninstall, pick, setMagnet, setStay, setLocked, setHidden,
    removeAll, removeSelected, submitText, cancelText,
  }
}
