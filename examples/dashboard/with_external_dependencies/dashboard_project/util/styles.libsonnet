local g = import './util/g.libsonnet';

{
    timeSeries: {
        local standardOptions = g.panel.timeSeries.standardOptions,

        default: {
            'fieldConfig': {
                'defaults': {
                  'color': {
                    'mode': 'palette-classic'
                  },
                  'custom': {
                      'axisPlacement': 'auto',
                      'barAlignment': 0,
                      'drawStyle': 'line',
                      'fillOpacity': 50,
                      'gradientMode': 'opacity',
                      'hideFrom': {
                        'legend': false,
                        'tooltip': false,
                        'viz': false
                      },
                      'lineInterpolation': 'smooth',
                      'lineStyle': {
                        'fill': 'solid'
                      },
                      'lineWidth': 1,
                      'pointSize': 5,
                      'scaleDistribution': {
                        'type': 'linear'
                      },
                      'showPoints': 'never',
                      'spanNulls': false,
                      'stacking': {
                        'group': 'A',
                        'mode': 'none'
                      },
                      'thresholdsStyle': {
                        'mode': 'off'
                      }
                    }
                }
            }
        },

        seconds:
            self.default
            + standardOptions.withUnit('s'),

        shorts:
            self.default
            + standardOptions.withUnit('short'),

    },
    heatMap: {
       local heatMap = g.panel.heatmap,

       tsBucket: {
        color: {"exponent": 0.5, 'cardColor': '#56A64B', 'colorScale': 'linear', 'colorScheme': 'interpolateReds', 'max': 100, 'min':0, 'mode': 'spectrum'},
       }
    }
}
