// DrawingLayer is a Lightweight Charts plugin that draws the user's drawings
// and the ruler on top of the price chart. How each shape looks is in
// shapes.js. Drawings live only in the browser's memory.
//
// A point is { time, price }: time is a bar's unix time, price is in ADA.
import { drawShape, drawRuler } from './shapes.js'

export class DrawingLayer {
  constructor() {
    this.drawings = [] // finished drawings
    this.draft = null // the drawing being placed right now
    this.ruler = null // the measurement, if any
    this.hidden = false
    this.t0 = 0 // time of the first bar
    this.step = 1 // seconds between bars
    this._chart = null
    this._series = null
    this._requestUpdate = null
    const layer = this
    this._view = { zOrder: () => 'top', renderer: () => ({ draw: (target) => layer._draw(target) }) }
  }

  // attached is called by the chart when the layer is added to a series.
  attached({ chart, series, requestUpdate }) {
    this._chart = chart
    this._series = series
    this._requestUpdate = requestUpdate
  }

  // detached is called by the chart when the layer is removed from a series.
  detached() {
    this._chart = this._series = this._requestUpdate = null
  }

  paneViews() {
    return [this._view]
  }

  // setBars tells the layer where the bars are, so a time can be turned into
  // an x position even when it is scrolled out of view.
  setBars(t0, step) {
    this.t0 = t0
    this.step = step || 1
  }

  // update asks the chart to redraw the layer.
  update() {
    this._requestUpdate?.()
  }

  // x is the horizontal position of a time, in pixels.
  x(time) {
    return this._chart?.timeScale().logicalToCoordinate((time - this.t0) / this.step) ?? null
  }

  // y is the vertical position of a price, in pixels.
  y(price) {
    return this._series?.priceToCoordinate(price) ?? null
  }

  // xy is the position of a point, or null when it cannot be placed.
  xy(p) {
    if (!p) return null
    const x = this.x(p.time)
    const y = this.y(p.price)
    return x == null || y == null ? null : { x, y }
  }

  _draw(target) {
    if (!this._chart || !this._series) return
    target.useMediaCoordinateSpace(({ context: ctx, mediaSize }) => {
      ctx.save()
      ctx.font = '12px Inter, system-ui, sans-serif'
      ctx.lineCap = 'round'
      if (!this.hidden) {
        for (const d of this.drawings) drawShape(ctx, this, d, mediaSize, false)
        if (this.draft && this.draft.type !== 'ruler') drawShape(ctx, this, this.draft, mediaSize, true)
        if (this.ruler) drawRuler(ctx, this, this.ruler)
      }
      if (this.draft?.type === 'ruler') drawRuler(ctx, this, this.draft)
      ctx.restore()
    })
  }
}
