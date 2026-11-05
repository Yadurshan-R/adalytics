// view.js: how the visitor moves around the chart. The chart can be dragged
// sideways and zoomed with the wheel; the toolbar's zoom buttons and reset
// use zoom and resetView.

// SCROLL lets the chart be dragged sideways (but not scrolled with the mouse
// wheel, which stays for scrolling the page).
export const SCROLL = { mouseWheel: false, pressedMouseMove: true, horzTouchDrag: true, vertTouchDrag: false }

// SCALE lets the chart zoom with the mouse wheel, a pinch or by dragging the axes.
export const SCALE = { mouseWheel: true, pinch: true, axisPressedMouseMove: true, axisDoubleClickReset: true }

// zoom shows fewer bars (factor below 1) or more (above 1), keeping the
// newest bar on the right.
export function zoom(chart, factor) {
  const ts = chart.timeScale()
  const r = ts.getVisibleLogicalRange()
  if (!r) return
  const width = Math.max(10, (r.to - r.from) * factor)
  ts.setVisibleLogicalRange({ from: r.to - width, to: r.to })
}

// resetView shows every bar loaded.
export const resetView = (chart) => chart.timeScale().fitContent()
