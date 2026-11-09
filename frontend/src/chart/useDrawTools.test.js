// Tests for useDrawTools, with a pretend manager in place of the library's
// DrawingManager (which needs a real chart in a browser): the drawings move
// with the series, one set per timeframe, the Measure tool, lock all,
// picking tools, the text box.
import { describe, it, expect, beforeEach } from 'vitest'
import { useDrawTools, MEASURE } from './useDrawTools.js'
import { forget } from './drawBoards.js'

// FakeManager keeps drawings in a list and sends the events the library sends.
class FakeManager {
  constructor(chart, series, opts) {
    Object.assign(this, { series, opts, list: [], tools: [], visible: true, destroyed: false, events: {} })
  }
  on(name, cb) { (this.events[name] ??= []).push(cb); return () => {} }
  emit(name, ...args) { (this.events[name] ?? []).forEach((cb) => cb(...args)) }
  exportJSON() { return JSON.stringify(this.list) }
  importJSON(json) { this.list = JSON.parse(json); this.emit('change', this.list) }
  destroy() { this.destroyed = true }
  setDrawingsVisible(v) { this.visible = v }
  setTool(kind) { this.tools.push(kind); if (!kind) this.emit('tool', null) }
  setMagnet(m) { this.opts.magnet = m }
  setStayInDrawingMode(on) { this.opts.stayInDrawingMode = on }
  drawings() { return this.list }
  get(id) { return this.list.find((d) => d.id === id) }
  add(d) { this.list.push(d); this.emit('add', d); this.emit('change', this.list); return d.id }
  update(d) { this.list = this.list.map((x) => (x.id === d.id ? d : x)); this.emit('change', this.list) }
  remove(id) { this.list = this.list.filter((d) => d.id !== id); this.emit('change', this.list) }
  clear() { this.list = []; this.emit('change', this.list) }
}

// setup gives the tools on a chart with two series, and the managers made.
function setup() {
  const made = []
  const state = { series: 'candles' }
  const el = { listeners: {}, addEventListener(n, f) { this.listeners[n] = f }, removeEventListener() {} }
  const t = useDrawTools({ chart: () => ({}), series: () => state.series, bars: () => [] }, (...a) => {
    made.push(new FakeManager(...a))
    return made.at(-1)
  })
  t.install(el)
  t.attach()
  return { t, made, state, el, dm: () => made.at(-1) }
}

describe('useDrawTools', () => {
  it('keeps the picked tool after a drawing by default (stay in drawing mode on)', () => {
    const { t, dm } = setup()
    expect(t.stay.value).toBe(true)
    expect(dm().opts.stayInDrawingMode).toBe(true)
    t.setStay(false)
    expect(dm().opts.stayInDrawingMode).toBe(false)
  })

  it('moves the drawings to the new series on a switch between candles and line', () => {
    const { t, made, state, dm } = setup()
    t.setMagnet('strong')
    t.setHidden(true)
    dm().add({ id: 'a', kind: 'trend-line' })
    dm().add({ id: 'b', kind: 'rectangle' })
    state.series = 'line'
    t.attach()
    expect(made).toHaveLength(2)
    expect(made[0].destroyed).toBe(true)
    expect(dm().series).toBe('line')
    expect(dm().list.map((d) => d.id)).toEqual(['a', 'b'])
    expect(dm().opts.magnet).toBe('strong')
    expect(dm().visible).toBe(false)
    expect(t.count.value).toBe(2)
    t.attach() // same series: nothing new
    expect(made).toHaveLength(2)
  })

  it('places a measurement that is not counted and goes on the next click', () => {
    const { t, el, dm } = setup()
    dm().add({ id: 'a', kind: 'trend-line' })
    t.pick(MEASURE)
    expect(t.tool.value).toBe(MEASURE)
    expect(dm().tools.at(-1)).toBe('date-and-price-range')
    dm().add({ id: 'm', kind: 'date-and-price-range' })
    expect(t.tool.value).toBe('cursor') // the tool is put down after one range
    expect(t.count.value).toBe(1)
    el.listeners.pointerdown()
    expect(dm().get('m')).toBeUndefined()
    expect(dm().get('a')).toBeTruthy()
  })

  it('locks every drawing, and new ones while lock is on', () => {
    const { t, dm } = setup()
    dm().add({ id: 'a', kind: 'ray' })
    t.setLocked(true)
    expect(dm().get('a').locked).toBe(true)
    dm().add({ id: 'b', kind: 'ray' })
    expect(dm().get('b').locked).toBe(true)
    t.selection.value = ['a', 'b']
    t.removeSelected()
    expect(dm().list).toHaveLength(2) // locked drawings stay
    t.setLocked(false)
    t.removeSelected()
    expect(dm().list).toHaveLength(0)
  })

  it('goes back to the cursor when a button picks the tool that is on, not from a menu', () => {
    const { t } = setup()
    t.pick('ray')
    t.pick('ray', false)
    expect(t.tool.value).toBe('ray')
    t.pick('ray')
    expect(t.tool.value).toBe('cursor')
  })

  it('writes typed text, and removes a new text left empty', () => {
    const { t, dm } = setup()
    dm().add({ id: 'x', kind: 'text' })
    dm().emit('textEdit', dm().get('x'), { x: 10, y: 20 })
    expect(t.textBox.value).toMatchObject({ id: 'x', x: 10, y: 20, fresh: true })
    t.submitText('  Breakout  ')
    expect(dm().get('x').text).toBe('Breakout')
    expect(t.textBox.value).toBeNull()
    dm().add({ id: 'y', kind: 'text' })
    dm().emit('textEdit', dm().get('y'), { x: 0, y: 0 })
    t.submitText('   ')
    expect(dm().get('y')).toBeUndefined()
  })

  describe('one set of drawings per timeframe', () => {
    beforeEach(forget)
    const ids = (dm) => dm.list.map((d) => d.id)

    it('hides a drawing on another timeframe and brings it back', () => {
      const { t, dm } = setup()
      t.show('night-ada|1h')
      dm().add({ id: 'a', kind: 'trend-line' })
      t.show('night-ada|4h')
      expect(ids(dm())).toEqual([])
      expect(t.count.value).toBe(0)
      dm().add({ id: 'b', kind: 'rectangle' })
      t.show('night-ada|1h')
      expect(ids(dm())).toEqual(['a'])
      expect(t.count.value).toBe(1)
      t.show('night-ada|4h')
      expect(ids(dm())).toEqual(['b'])
    })

    it('keeps them only until deleted, and never keeps the measurement', () => {
      const { t, dm } = setup()
      t.show('night-ada|1h')
      dm().add({ id: 'a', kind: 'trend-line' })
      t.removeAll()
      t.show('night-ada|4h')
      t.show('night-ada|1h')
      expect(ids(dm())).toEqual([])
      t.pick(MEASURE)
      dm().add({ id: 'm', kind: 'date-and-price-range' })
      t.show('night-ada|4h')
      t.show('night-ada|1h')
      expect(ids(dm())).toEqual([])
    })

    it('keeps them when the chart is left and built again (until the page reloads)', () => {
      const first = setup()
      first.t.show('night-ada|1h')
      first.dm().add({ id: 'a', kind: 'trend-line' })
      first.t.uninstall()
      const made = []
      const t = useDrawTools({ chart: () => ({}), series: () => 'candles', bars: () => [] }, (...a) => made.at(made.push(new FakeManager(...a)) - 1))
      t.show('night-ada|1h') // the page sets the timeframe before the chart is built
      t.attach()
      expect(ids(made[0])).toEqual(['a'])
    })
  })
})
