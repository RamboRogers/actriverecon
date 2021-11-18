// https://observablehq.com/@d3/indented-tree@335
export default function define(runtime, observer) {
  const main = runtime.module();
  const fileAttachments = new Map([["flare-2.json",new URL("./files/e65374209781891f37dea1e7a6e1c5e020a3009b8aedf113b4c80942018887a1176ad4945cf14444603ff91d3da371b3b0d72419fa8d2ee0f6e815732475d5de",import.meta.url)]]);
  main.builtin("FileAttachment", runtime.fileAttachments(name => fileAttachments.get(name)));
  main.variable(observer()).define(["md"], function(md){return(
md`# Indented Tree

A tabular layout for hierarchical data, indented trees allow one or more columns of values to be shown alongside indented names.`
)});
  main.variable(observer("chart")).define("chart", ["root","d3","nodeSize","width","columns"], function(root,d3,nodeSize,width,columns)
{
  const nodes = root.descendants();

  const svg = d3.create("svg")
      .attr("viewBox", [-nodeSize / 2, -nodeSize * 3 / 2, width, (nodes.length + 1) * nodeSize])
      .attr("font-family", "sans-serif")
      .attr("font-size", 10)
      .style("overflow", "visible");

  const link = svg.append("g")
      .attr("fill", "none")
      .attr("stroke", "#999")
    .selectAll("path")
    .data(root.links())
    .join("path")
      .attr("d", d => `
        M${d.source.depth * nodeSize},${d.source.index * nodeSize}
        V${d.target.index * nodeSize}
        h${nodeSize}
      `);

  const node = svg.append("g")
    .selectAll("g")
    .data(nodes)
    .join("g")
      .attr("transform", d => `translate(0,${d.index * nodeSize})`);

  node.append("circle")
      .attr("cx", d => d.depth * nodeSize)
      .attr("r", 2.5)
      .attr("fill", d => d.children ? null : "#999");

  node.append("text")
      .attr("dy", "0.32em")
      .attr("x", d => d.depth * nodeSize + 6)
      .text(d => d.data.name);

  node.append("title")
      .text(d => d.ancestors().reverse().map(d => d.data.name).join("/"));

  for (const {label, value, format, x} of columns) {
    svg.append("text")
        .attr("dy", "0.32em")
        .attr("y", -nodeSize)
        .attr("x", x)
        .attr("text-anchor", "end")
        .attr("font-weight", "bold")
        .text(label);

    node.append("text")
        .attr("dy", "0.32em")
        .attr("x", x)
        .attr("text-anchor", "end")
        .attr("fill", d => d.children ? null : "#555")
      .data(root.copy().sum(value).descendants())
        .text(d => format(d.value, d));
  }

  return svg.node();
}
);
  main.variable(observer("columns")).define("columns", ["format"], function(format){return(
[
  {
    label: "Size", 
    value: d => d.value, 
    format, 
    x: 280
  },
  {
    label: "Count", 
    value: d => d.children ? 0 : 1, 
    format: (value, d) => d.children ? format(value) : "-", 
    x: 340
  }
]
)});
  main.variable(observer("format")).define("format", ["d3"], function(d3){return(
d3.format(",")
)});
  main.variable(observer("data")).define("data", ["FileAttachment"], function(FileAttachment){return(
FileAttachment("flare-2.json").json()
)});
  main.variable(observer("root")).define("root", ["d3","data"], function(d3,data)
{ let i = 0; return d3.hierarchy(data).eachBefore(d => d.index = i++); }
);
  main.variable(observer("nodeSize")).define("nodeSize", function(){return(
17
)});
  main.variable(observer("d3")).define("d3", ["require"], function(require){return(
require("d3@6")
)});
  return main;
}
