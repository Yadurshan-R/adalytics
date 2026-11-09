// The icons of the drawing toolbar, drawn for this app in the thin-line style
// of TradingView's toolbar. Each icon is an SVG path on a 20 × 20 grid
// (`icon`), with small circles for the points the visitor clicks (`dots`).
// ToolIcon.vue draws them.

// A circle as a path, so every icon is one path: centre (x, y), radius r.
const ring = (x, y, r) => `M${x + r} ${y}a${r} ${r} 0 1 1-${2 * r} 0a${r} ${r} 0 1 1 ${2 * r} 0`

// ICONS has one icon per drawing tool (by its kind) and per toolbar action.
export const ICONS = {
  // Cursor
  cursor: { icon: 'M10 2.5v5.5M10 12v5.5M2.5 10H8M12 10h5.5' },

  // Lines and channels
  'trend-line': { icon: 'M5.6 14.4 14.4 5.6', dots: [[4.5, 15.5], [15.5, 4.5]] },
  ray: { icon: 'M5.6 14.4 18 2', dots: [[4.5, 15.5], [10.5, 9.5]] },
  'horizontal-line': { icon: 'M2 10h6.4M11.6 10H18', dots: [[10, 10]] },
  'horizontal-ray': { icon: 'M7.6 10H18', dots: [[6, 10]] },
  'vertical-line': { icon: 'M10 2v6.4M10 11.6V18', dots: [[10, 10]] },
  'parallel-channel': { icon: 'M4.1 10.9 10.9 4.1M8 17 15.9 9.1', dots: [[3, 12], [12, 3], [17, 8]] },

  // Projection and measurer
  'long-position': { icon: 'M4.6 4H18M4.6 10H18M4.6 16H18M7 10l9-5', dots: [[3, 4], [3, 10], [3, 16]] },
  'short-position': { icon: 'M4.6 4H18M4.6 10H18M4.6 16H18M7 10l9 5', dots: [[3, 4], [3, 10], [3, 16]] },
  'date-and-price-range': { icon: 'M3 3h14v14H3zM10 6v8M8 8l2-2 2 2M6 10h8M12 8l2 2-2 2' },

  // Brushes, arrows and shapes
  brush: { icon: 'M3 15.5c2.5-.5 3-6.5 6-6.5s2.5 5.5 5.5 3.5S17 6 17 6' },
  arrow: { icon: 'M4 16 15.5 4.5M9.5 4.5h6v6' },
  rectangle: { icon: 'M5.1 5h9.8M16 6.1v7.8M14.9 15H5.1M4 13.9V6.1', dots: [[4, 5], [16, 5], [16, 15], [4, 15]] },

  // Text and notes
  text: { icon: 'M4 5.5V3.5h12v2M10 3.5v13M7.5 16.5h5' },
  'price-label': { icon: 'M3 6h10.5L17 10l-3.5 4H3zM6 10h5' },

  // Actions
  measure: { icon: 'M3 14 14 3l3 3L6 17zM6.5 10.5l1.5 1.5M9 8l1.5 1.5M11.5 5.5l1.5 1.5' },
  zoomIn: { icon: `${ring(8.5, 8.5, 5.5)}M12.5 12.5 17 17M6 8.5h5M8.5 6v5` },
  zoomOut: { icon: `${ring(8.5, 8.5, 5.5)}M12.5 12.5 17 17M6 8.5h5` },
  reset: { icon: 'M3 7V3h4M17 7V3h-4M3 13v4h4M17 13v4h-4' },
  weak: { icon: 'M5 3v7a5 5 0 0 0 10 0V3M5 6.5h3.5M11.5 6.5H15M8.5 3v7a1.5 1.5 0 0 0 3 0V3' },
  strong: { icon: 'M5 2v6.5a5 5 0 0 0 10 0V2M5 5h3.5M11.5 5H15M8.5 2v6.5a1.5 1.5 0 0 0 3 0V2M7 15.5l1 1.5-1 1.5M13 15.5l1 1.5-1 1.5' },
  stay: { icon: 'M3 15l.8-3.2 7-7 2.4 2.4-7 7zM9.4 6.2l2.4 2.4M12 13h6v4.5h-6zM13.5 13v-1.2a1.5 1.5 0 0 1 3 0V13' },
  lock: { icon: 'M4.5 9h11v8h-11zM7 9V6.5a3 3 0 0 1 6 0V9' },
  unlock: { icon: 'M4.5 9h11v8h-11zM7 9V6.5a3 3 0 0 1 5.8-1.1' },
  eye: { icon: `M2 10s3-5.5 8-5.5S18 10 18 10s-3 5.5-8 5.5S2 10 2 10z${ring(10, 10, 2.5)}` },
  eyeOff: { icon: `M2 10s3-5.5 8-5.5S18 10 18 10s-3 5.5-8 5.5S2 10 2 10z${ring(10, 10, 2.5)}M3 3l14 14` },
  indicators: { icon: 'M2.5 15.5 7 10l3.5 3 7-8.5M2.5 17.5h15' },
  trash: { icon: 'M3.5 5.5h13M8 5.5V3.5h4v2M5.5 5.5l1 11h7l1-11M8.5 8.5v5M11.5 8.5v5' },
}
