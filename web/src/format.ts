import type { Point } from "./types";

export function latest(points: Point[]): number | null {
  return points.length ? points[points.length - 1][1] : null;
}

export function formatPct(v: number | null): string {
  return v === null ? "--" : `${Math.round(v)}%`;
}

const rateUnits = ["B/s", "KB/s", "MB/s", "GB/s"];

export function formatRate(v: number | null): string {
  if (v === null) return "--";
  let i = 0;
  while (v >= 1000 && i < rateUnits.length - 1) {
    v /= 1000;
    i++;
  }
  return `${v >= 10 || i === 0 ? Math.round(v) : v.toFixed(1)} ${rateUnits[i]}`;
}

export function formatClock(d: Date): string {
  return d.toLocaleTimeString([], { hour12: false });
}
