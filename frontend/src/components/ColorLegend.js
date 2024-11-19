import React, { useRef, useEffect } from "react";
import { select } from "d3-selection";
import { scaleLinear } from "d3-scale";
import { axisBottom } from "d3-axis";
import { format } from "d3-format";

const ColorLegend = ({ colorScale, title, ticks }) => {
  const legendRef = useRef();

  useEffect(() => {
    const legend = select(legendRef.current);

    legend.selectAll("*").remove();

    const svg = legend.append("svg").attr("width", 322).attr("height", 60);

    const gradient = svg
      .append("defs")
      .append("linearGradient")
      .attr("id", "gradient")
      .attr("x1", "0%")
      .attr("y1", "100%")
      .attr("x2", "100%")
      .attr("y2", "100%")
      .attr("spreadMethod", "pad");

    gradient
      .selectAll("stop")
      .data(
        colorScale.ticks(ticks).map((t, i, n) => ({
          offset: `${(100 * i) / n.length}%`,
          color: colorScale(t),
        })),
      )
      .enter()
      .append("stop")
      .attr("offset", (d) => d.offset)
      .attr("stop-color", (d) => d.color);

    svg
      .append("rect")
      .attr("width", 300)
      .attr("height", 8)
      .style("fill", "url(#gradient)")
      .attr("transform", "translate(10, 20)")
      .attr("stroke", "white")
      .attr("stroke-width", 1);

    svg
      .append("g")
      .attr("transform", "translate(10, 28)")
      .call(
        axisBottom(scaleLinear().range([0, 300]).domain([0, 2000]))
          .ticks(ticks)
          .tickFormat(format("2")),
      )
      .call((g) => g.select(".domain").remove())
      .call((g) => g.selectAll(".tick line").remove())
      .selectAll("text")
      .attr("fill", "white");

    svg
      .append("text")
      .attr("x", 0)
      .attr("y", 10)
      .attr("text-anchor", "start")
      .attr("font-weight", "bold")
      .attr("fill", "white")
      .text(title);
    svg
      .append("text")
      .attr("x", 310)
      .attr("y", 58)
      .attr("text-anchor", "end")
      .attr("font-size", "10px")
      .attr("fill", "white")
      .html(
        'nanoWatts/cm<tspan dy="-0.5em" font-size="8px">2</tspan><tspan dy="0.4em"' +
          ' font-size="10px">/sr</tspan>',
      );
  }, [colorScale, title, ticks]);

  return (
    <div
      ref={legendRef}
      style={{
        position: "absolute",
        bottom: "20px",
        right: "20px",
        zIndex: 10,
      }}
    ></div>
  );
};

export default ColorLegend;
