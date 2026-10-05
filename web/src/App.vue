<script setup lang="ts">
import { computed } from "vue";
import ClusterHeader from "./components/ClusterHeader.vue";
import NodeTile from "./components/NodeTile.vue";
import { formatClock } from "./format";
import { useNodes } from "./composables/useNodes";

const WINDOW_SECONDS = 15 * 60;
const COLUMNS = 4;

const { snapshot, lastOk, now, stale } = useNodes();

const end = computed(() => (snapshot.value ? Date.parse(snapshot.value.updated) / 1000 : now.value / 1000));
const rows = computed(() => Math.max(1, Math.ceil((snapshot.value?.nodes.length ?? 0) / COLUMNS)));
</script>

<template>
  <main class="board" :class="{ 'is-stale': stale }">
    <ClusterHeader :snapshot="snapshot" :now="now" :last-ok="lastOk" />
    <section class="grid" :style="{ gridTemplateRows: `repeat(${rows}, minmax(0, 1fr))` }">
      <NodeTile v-for="node in snapshot?.nodes ?? []" :key="node.name" :node="node" :end="end" :window="WINDOW_SECONDS" />
    </section>
    <div v-if="stale" class="stale">
      DATA STALE
      <small>{{ lastOk === null ? "no data received" : `last update ${formatClock(new Date(lastOk))}` }}</small>
    </div>
  </main>
</template>

<style scoped>
.board {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  height: 100vh;
  padding: 0.75rem;
  box-sizing: border-box;
  overflow: hidden;
}
.grid {
  flex: 1;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.75rem;
  min-height: 0;
}
.is-stale > :not(.stale) {
  opacity: 0.25;
  filter: grayscale(1);
}
.stale {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  font-size: 4rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  color: var(--warn);
}
.stale small {
  font-size: 1.5rem;
  font-weight: 600;
  letter-spacing: 0.04em;
}
</style>
