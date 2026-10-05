<script setup lang="ts">
import { computed } from "vue";
import Sparkline from "./Sparkline.vue";
import { formatPct, formatRate, latest } from "../format";
import type { ClusterNode, Point } from "../types";

const props = defineProps<{ node: ClusterNode; end: number; window: number }>();

const MEM_THRESHOLD = 90;
const MB = 1_000_000;

interface Row {
  label: string;
  points: Point[];
  value: string;
  max?: number;
  floor?: number;
  breached?: boolean;
}

const rows = computed<Row[]>(() => {
  const s = props.node.series;
  const mem = latest(s.mem);
  return [
    { label: "CPU", points: s.cpu, value: formatPct(latest(s.cpu)), max: 100 },
    { label: "MEM", points: s.mem, value: formatPct(mem), max: 100, breached: mem !== null && mem > MEM_THRESHOLD },
    { label: "DSK", points: s.disk, value: formatRate(latest(s.disk)), floor: MB },
    { label: "NET", points: s.net, value: formatRate(latest(s.net)), floor: MB },
  ];
});

const ok = computed(() => props.node.health === "ok");
const outage = computed(() => (!props.node.ready ? "NOT READY" : "NO METRICS"));
</script>

<template>
  <article class="tile" :class="ok ? 'is-ok' : 'is-bad'">
    <header class="strip">
      <span class="icon" aria-hidden="true">{{ ok ? "✔" : "✖" }}</span>
      <h2 class="name">{{ node.name }}</h2>
      <span v-if="node.role === 'control-plane'" class="badge">cp</span>
      <span v-if="node.arch" class="badge">{{ node.arch }}</span>
      <span v-if="node.cordoned" class="badge badge-warn">CORDONED</span>
      <span class="state">{{ ok ? "OK" : "DOWN" }}</span>
    </header>

    <div v-if="node.exporter" class="rows">
      <div v-for="row in rows" :key="row.label" class="row">
        <span class="label">{{ row.label }}</span>
        <div class="spark">
          <Sparkline :points="row.points" :end="end" :window="window" :max="row.max" :floor="row.floor" />
        </div>
        <span class="value" :class="{ breached: row.breached }">{{ row.value }}</span>
      </div>
    </div>
    <div v-else class="outage">{{ outage }}</div>

    <footer class="foot">
      <span>
        ROOT
        <span :class="{ breached: node.rootFsPct !== null && node.rootFsPct > 90 }">{{ formatPct(node.rootFsPct) }}</span>
      </span>
      <span v-if="node.reasons.length" class="reasons">{{ node.reasons.join(" · ") }}</span>
    </footer>
  </article>
</template>

<style scoped>
.tile {
  display: flex;
  flex-direction: column;
  min-height: 0;
  border: 0.25rem solid var(--health);
  border-radius: 0.75rem;
  background: var(--surface);
  overflow: hidden;
}
.is-ok {
  --health: var(--ok);
}
.is-bad {
  --health: var(--bad);
  animation: pulse 2s ease-in-out infinite;
}
@keyframes pulse {
  50% {
    box-shadow: 0 0 1.5rem 0.25rem var(--bad);
  }
}
@media (prefers-reduced-motion: reduce) {
  .is-bad {
    animation: none;
  }
}

.strip {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.4rem 0.9rem;
  background: var(--health);
  color: var(--on-health);
}
.icon {
  font-size: 1.5rem;
  font-weight: 700;
}
.name {
  margin: 0;
  font-size: 1.75rem;
  font-weight: 700;
  letter-spacing: 0.01em;
}
.badge {
  padding: 0.05rem 0.45rem;
  border: 2px solid currentColor;
  border-radius: 0.35rem;
  font-size: 0.85rem;
  font-weight: 700;
  text-transform: uppercase;
}
.badge-warn {
  background: var(--on-health);
  color: var(--health);
  border-color: var(--on-health);
}
.state {
  margin-left: auto;
  font-size: 1.25rem;
  font-weight: 800;
  letter-spacing: 0.08em;
}

.rows {
  flex: 1;
  display: grid;
  grid-auto-rows: 1fr;
  gap: 0.35rem;
  padding: 0.6rem 0.9rem 0.2rem;
  min-height: 0;
}
.row {
  display: grid;
  grid-template-columns: 3rem 1fr 8.5rem;
  align-items: center;
  gap: 0.75rem;
  min-height: 0;
}
.label {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-muted);
  letter-spacing: 0.06em;
}
.spark {
  height: 100%;
  min-height: 0;
}
.value {
  text-align: right;
  white-space: nowrap;
  font-size: 1.4rem;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  color: var(--text);
}
.breached {
  color: var(--bad);
  font-weight: 800;
}

.outage {
  flex: 1;
  display: grid;
  place-items: center;
  font-size: 2.25rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  color: var(--bad);
}

.foot {
  display: flex;
  gap: 1rem;
  padding: 0.3rem 0.9rem 0.5rem;
  font-size: 1rem;
  color: var(--text-muted);
  font-variant-numeric: tabular-nums;
}
.reasons {
  margin-left: auto;
  color: var(--bad);
  font-weight: 700;
  text-transform: uppercase;
}
</style>
