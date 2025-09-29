local g = import './g.libsonnet';

{
  timeSeries: {
    local timeSeries = g.panel.timeSeries,
    local fieldConfig = timeSeries.fieldConfig,
    local defaults = timeSeries.fieldConfig.defaults,
    local custom = timeSeries.fieldConfig.defaults.custom,
    local override = timeSeries.fieldConfig.overrides,
    local options = timeSeries.options,

    base(title, targets, datasource, fieldConfig = {}, grid):
      timeSeries.new(title)
      + timeSeries.withTargets(targets)
      + datasource
      + fieldConfig
      + grid,

    clickhouseTimeSeries(title, targets, datasource, fieldConfig = {}, grid, interval = '1m'):
        self.base(title, targets, datasource, fieldConfig, grid)
        + { 'interval': interval },
  },

  heatMap: {
    local heatMap = g.panel.heatmap,
    local options = heatMap.options,
    local defaultColor = {"exponent": 0.5,"fill": "dark-orange","reverse": false,"scheme": "Oranges","steps": 64},

    base(title, targets, datasource, grid = {}):
        heatMap.new(title)
        + heatMap.withTargets(targets)
        + datasource
        + grid,

    tsBucketsHeatMap(title, targets, datasource, color = defaultColor, grid, interval = '1m', showLegend = 1):
        self.base(title, targets, datasource, grid)
        + { 'color': color }
        + { 'cards': {'cardRound': 3 }}
        + { 'interval': interval }
        + { 'legend': { 'show': showLegend} }
        + {'dataFormat': 'tsbuckets'}
        + { 'yBucketBound': 'middle' }
  }
}
