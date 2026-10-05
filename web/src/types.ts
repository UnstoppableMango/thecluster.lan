/** [unix seconds, value] */
export type Point = [number, number];

export type Health = "ok" | "bad";

export interface NodeSeries {
  cpu: Point[];
  mem: Point[];
  disk: Point[];
  net: Point[];
}

export interface ClusterNode {
  name: string;
  role: string;
  arch?: string;
  health: Health;
  reasons: string[];
  ready: boolean;
  cordoned: boolean;
  exporter: boolean;
  rootFsPct: number | null;
  series: NodeSeries;
}

export interface Snapshot {
  updated: string;
  cluster: { cpuPct: number | null; memPct: number | null };
  nodes: ClusterNode[];
}
