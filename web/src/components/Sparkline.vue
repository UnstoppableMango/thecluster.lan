<script setup lang="ts">
import { computed } from "vue";
import type { Point } from "../types";

const props = defineProps<{
  points: Point[];
  /** Window end, unix seconds. */
  end: number;
  window: number;
  /** Fixed y maximum (e.g. 100 for percentages). Omit to auto-scale. */
  max?: number;
  /** Lowest auto-scaled maximum, so idle noise does not fill the chart. */
  floor?: number;
}>();

const W = 100;
const H = 40;

const yMax = computed(() => {
  if (props.max !== undefined) return props.max;
  const peak = Math.max(0, ...props.points.map((p) => p[1]));
  return Math.max(peak * 1.1, props.floor ?? 0) || 1;
});

const coords = computed(() => {
  const start = props.end - props.window;
  return props.points.map(([t, v]) => {
    const x = ((t - start) / props.window) * W;
    const y = H - (Math.min(v, yMax.value) / yMax.value) * H;
    return `${x.toFixed(2)},${y.toFixed(2)}`;
  });
});

const line = computed(() => coords.value.join(" "));
const area = computed(() => {
  if (coords.value.length < 2) return "";
  const firstX = coords.value[0].split(",")[0];
  const lastX = coords.value[coords.value.length - 1].split(",")[0];
  return `${firstX},${H} ${line.value} ${lastX},${H}`;
});
</script>

<template>
  <svg :viewBox="`0 0 ${W} ${H}`" preserveAspectRatio="none" class="block h-full w-full" aria-hidden="true">
    <line x1="0" :y1="H" :x2="W" :y2="H" class="baseline" vector-effect="non-scaling-stroke" />
    <polygon v-if="area" :points="area" class="area" />
    <polyline :points="line" class="line" vector-effect="non-scaling-stroke" />
  </svg>
</template>

<style scoped>
.baseline {
  stroke: var(--grid);
  stroke-width: 1;
}
.area {
  fill: var(--spark);
  opacity: 0.18;
}
.line {
  fill: none;
  stroke: var(--spark);
  stroke-width: 2;
  stroke-linejoin: round;
  stroke-linecap: round;
}
</style>
