<template>
  <div class="flex flex-col items-center">
    <div ref="chartRef" class="w-42 h-42"></div>
    <span class="text-center text-base-content/50">{{ title }}</span>
  </div>
</template>

<script setup lang="ts">
import * as echarts from 'echarts/core';
import { GaugeChart, type GaugeSeriesOption } from 'echarts/charts';
import { CanvasRenderer } from 'echarts/renderers';

echarts.use([GaugeChart, CanvasRenderer]);

type EChartsOption = echarts.ComposeOption<GaugeSeriesOption>;

const props = defineProps<{
  title: string;
  value: number;
}>();

const chartRef = ref<HTMLElement>();
const chart = ref<echarts.ECharts>();
onMounted(() => {
  if (chartRef.value) {
    init();
  }
});

onUnmounted(() => {
  chart.value?.dispose();
});

watch(() => props.value, () => {
  init();
});

const init = function () {
  if (chart.value) {
    chart.value.dispose();
  }
  chart.value = echarts.init(chartRef.value);
  var option: EChartsOption;
  const value = +props.value.toFixed(2);
  option = {
    series: [
      {
        type: 'gauge',
        startAngle: 90,
        endAngle: -270, // 形成一个完整的圆
        pointer: { show: false }, // 隐藏指针
        progress: {
          show: true,
          overlap: false,
          roundCap: true, // 核心：开启圆角
          clip: false,
          itemStyle: {
            color: '#3b82f6' // 蓝色进度条
          }
        },
        axisLine: {
          lineStyle: {
            width: 12, // 环的宽度
            color: [[1, '#f0f5ff']] // 底色/轨道颜色
          }
        },
        splitLine: { show: false }, // 隐藏刻度线
        axisTick: { show: false },  // 隐藏小刻度
        axisLabel: { show: false }, // 隐藏数字标签
        data: [{ value }],
        detail: {
          borderRadius: 8,
          offsetCenter: [0, 0],
          formatter: function (value) {
            return `${value}%`;
          },
          fontSize: 24,
        }
      }
    ]
  };
  chart.value.setOption(option);
};


</script>
