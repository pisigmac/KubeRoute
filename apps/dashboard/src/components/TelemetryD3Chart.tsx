"use client";

import { useEffect, useRef, useState } from "react";
import * as d3 from "d3";

interface DataPoint {
  time: string;
  rps: number;
  savings: number;
}

export default function TelemetryD3Chart() {
  const containerRef = useRef<HTMLDivElement>(null);
  const svgRef = useRef<SVGSVGElement>(null);

  const [hoverData, setHoverData] = useState<{
    time: string;
    rps: number;
    savings: number;
    x: number;
    y: number;
  } | null>(null);

  useEffect(() => {
    if (!containerRef.current || !svgRef.current) return;

    // Generate mock 24-point time series data
    const data: DataPoint[] = Array.from({ length: 24 }, (_, i) => {
      const hour = (i + 12) % 24;
      const timeStr = `${hour.toString().padStart(2, "0")}:00`;
      const baseRps = 1200 + Math.sin(i * 0.5) * 600 + Math.random() * 80;
      const baseSavings = 35 + Math.cos(i * 0.4) * 15 + Math.random() * 5;
      return {
        time: timeStr,
        rps: Math.round(baseRps),
        savings: Math.round(baseSavings * 10) / 10,
      };
    });

    const margin = { top: 30, right: 30, bottom: 40, left: 50 };
    const width = containerRef.current.clientWidth - margin.left - margin.right;
    const height = 240 - margin.top - margin.bottom;

    const svg = d3.select(svgRef.current);
    svg.selectAll("*").remove();

    svg
      .attr("width", width + margin.left + margin.right)
      .attr("height", height + margin.top + margin.bottom);

    const g = svg
      .append("g")
      .attr("transform", `translate(${margin.left},${margin.top})`);

    // Scales
    const xScale = d3
      .scalePoint()
      .domain(data.map((d) => d.time))
      .range([0, width]);

    const rpsMax = d3.max(data, (d) => d.rps) || 2000;
    const yScaleRps = d3
      .scaleLinear()
      .domain([0, rpsMax * 1.15])
      .range([height, 0]);

    // Define Gradient Fill for D3 Area
    const defs = svg.append("defs");

    const areaGradient = defs
      .append("linearGradient")
      .attr("id", "blueAreaGradient")
      .attr("x1", "0%")
      .attr("y1", "0%")
      .attr("x2", "0%")
      .attr("y2", "100%");

    areaGradient
      .append("stop")
      .attr("offset", "0%")
      .attr("stop-color", "#3B82F6")
      .attr("stop-opacity", 0.45);

    areaGradient
      .append("stop")
      .attr("offset", "100%")
      .attr("stop-color", "#3B82F6")
      .attr("stop-opacity", 0.0);

    // Grid Lines
    const yGrid = d3.axisLeft(yScaleRps).ticks(4).tickSize(-width).tickFormat(() => "");
    g.append("g")
      .attr("class", "grid")
      .call(yGrid)
      .selectAll("line")
      .attr("stroke", "#1E293B")
      .attr("stroke-dasharray", "3,3");

    g.selectAll(".grid .domain").remove();

    // D3 Area Generator for Traffic Surge Forecast
    const area = d3
      .area<DataPoint>()
      .x((d) => xScale(d.time) || 0)
      .y0(height)
      .y1((d) => yScaleRps(d.rps))
      .curve(d3.curveMonotoneX);

    g.append("path")
      .datum(data)
      .attr("fill", "url(#blueAreaGradient)")
      .attr("d", area);

    // D3 Line Generator for Traffic Line
    const lineRps = d3
      .line<DataPoint>()
      .x((d) => xScale(d.time) || 0)
      .y((d) => yScaleRps(d.rps))
      .curve(d3.curveMonotoneX);

    g.append("path")
      .datum(data)
      .attr("fill", "none")
      .attr("stroke", "#3B82F6")
      .attr("stroke-width", 3)
      .attr("d", lineRps);

    // D3 Line Generator for Spot Cost Savings
    const yScaleSavings = d3
      .scaleLinear()
      .domain([0, 60])
      .range([height, 0]);

    const lineSavings = d3
      .line<DataPoint>()
      .x((d) => xScale(d.time) || 0)
      .y((d) => yScaleSavings(d.savings))
      .curve(d3.curveMonotoneX);

    g.append("path")
      .datum(data)
      .attr("fill", "none")
      .attr("stroke", "#10B981")
      .attr("stroke-width", 2.5)
      .attr("stroke-dasharray", "4,2")
      .attr("d", lineSavings);

    // X Axis
    const xAxis = d3.axisBottom(xScale).tickValues(
      data.filter((_, idx) => idx % 4 === 0).map((d) => d.time)
    );

    g.append("g")
      .attr("transform", `translate(0,${height})`)
      .call(xAxis)
      .attr("color", "#64748B")
      .selectAll("text")
      .attr("font-family", "monospace")
      .attr("font-size", "10px");

    // Y Axis (RPS)
    const yAxis = d3.axisLeft(yScaleRps).ticks(4);
    g.append("g")
      .call(yAxis)
      .attr("color", "#64748B")
      .selectAll("text")
      .attr("font-family", "monospace")
      .attr("font-size", "10px");

    // Interactive Hover Overlay Line
    const focusLine = g
      .append("line")
      .attr("stroke", "#94A3B8")
      .attr("stroke-width", 1)
      .attr("stroke-dasharray", "3,3")
      .style("opacity", 0);

    const focusCircleRps = g
      .append("circle")
      .attr("r", 5)
      .attr("fill", "#FFFFFF")
      .attr("stroke", "#3B82F6")
      .attr("stroke-width", 2.5)
      .style("opacity", 0);

    // Overlay Rect for Mouse Events
    g.append("rect")
      .attr("width", width)
      .attr("height", height)
      .attr("fill", "none")
      .attr("pointer-events", "all")
      .on("mousemove", (event) => {
        const [mouseX] = d3.pointer(event);
        const eachBand = width / (data.length - 1);
        const index = Math.min(
          data.length - 1,
          Math.max(0, Math.round(mouseX / eachBand))
        );
        const d = data[index];
        const xPos = xScale(d.time) || 0;
        const yPos = yScaleRps(d.rps);

        focusLine.attr("x1", xPos).attr("x2", xPos).attr("y1", 0).attr("y2", height).style("opacity", 1);
        focusCircleRps.attr("cx", xPos).attr("cy", yPos).style("opacity", 1);

        setHoverData({
          time: d.time,
          rps: d.rps,
          savings: d.savings,
          x: xPos + margin.left,
          y: yPos + margin.top,
        });
      })
      .on("mouseleave", () => {
        focusLine.style("opacity", 0);
        focusCircleRps.style("opacity", 0);
        setHoverData(null);
      });
  }, []);

  return (
    <div ref={containerRef} className="w-full bg-[#0E0F17] rounded-2xl border border-gray-800 p-5 shadow-2xl relative overflow-hidden">
      <div className="flex flex-wrap justify-between items-center mb-2 px-1 text-xs font-mono">
        <div className="flex items-center gap-2">
          <span className="w-2.5 h-2.5 rounded-full bg-blue-500 animate-pulse" />
          <span className="text-gray-200 font-bold">D3.js Interactive Real-Time Telemetry Chart</span>
        </div>
        <div className="flex items-center gap-4 text-[11px]">
          <span className="text-blue-400 font-bold">■ Traffic Surge Forecast (RPS)</span>
          <span className="text-emerald-400 font-bold">-- Spot Cost Savings (%)</span>
        </div>
      </div>

      {/* SVG Canvas for D3 Rendering */}
      <svg ref={svgRef} className="w-full h-[240px] overflow-visible" />

      {/* Interactive Tooltip on Hover */}
      {hoverData && (
        <div
          className="absolute pointer-events-none bg-gray-900/95 text-white text-[11px] font-mono px-3 py-2 rounded-lg border border-gray-700 shadow-xl backdrop-blur-md z-10 transition-all duration-75"
          style={{
            left: `${hoverData.x + 10}px`,
            top: `${hoverData.y - 40}px`,
          }}
        >
          <p className="text-gray-400 font-bold border-b border-gray-800 pb-1 mb-1">Time: {hoverData.time}</p>
          <p className="text-blue-400">Traffic: <span className="text-white font-bold">{hoverData.rps} RPS</span></p>
          <p className="text-emerald-400">Spot Savings: <span className="text-white font-bold">{hoverData.savings}%</span></p>
        </div>
      )}

      <div className="flex justify-between items-center mt-3 pt-2 border-t border-gray-800/80 text-[11px] font-mono text-gray-400">
        <span>Engine: D3.js v7 Data-Driven Document Renderer</span>
        <span className="text-emerald-400 font-semibold">Hover to Inspect Telemetry</span>
      </div>
    </div>
  );
}
