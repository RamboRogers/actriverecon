// https://observablehq.com/@d3/tidy-tree@170
export default function define(runtime, observer) {
  const main = runtime.module();
  const fileAttachments = new Map([["flare-2.json",new URL("./files/e65374209781891f37dea1e7a6e1c5e020a3009b8aedf113b4c80942018887a1176ad4945cf14444603ff91d3da371b3b0d72419fa8d2ee0f6e815732475d5de",import.meta.url)]]);
  main.builtin("FileAttachment", runtime.fileAttachments(name => fileAttachments.get(name)));
  main.variable(observer()).define(["md"], function(md){return(
md`# Tidy Tree

D3’s [tree layout](https://github.com/d3/d3-hierarchy/blob/master/README.md#tree) implements the [Reingold–Tilford “tidy” algorithm](http://reingold.co/tidier-drawings.pdf) for constructing hierarchical node-link diagrams, improved to run in linear time by [Buchheim *et al.*](http://dirk.jivas.de/papers/buchheim02improving.pdf) Tidy trees are typically more compact than [cluster dendrograms](/@d3/cluster-dendrogram), which place all leaves at the same level. See also the [radial variant](/@d3/radial-tidy-tree).`
)});
  main.variable(observer("chart")).define("chart", ["tree","data","d3","width"], function(tree,data,d3,width)
{
  const root = tree(data);

  let x0 = Infinity;
  let x1 = -x0;
  root.each(d => {
    if (d.x > x1) x1 = d.x;
    if (d.x < x0) x0 = d.x;
  });

  const svg = d3.create("svg")
      .attr("viewBox", [0, 0, width, x1 - x0 + root.dx * 2]);
  
  const g = svg.append("g")
      .attr("font-family", "sans-serif")
      .attr("font-size", 10)
      .attr("transform", `translate(${root.dy / 3},${root.dx - x0})`);
    
  const link = g.append("g")
    .attr("fill", "none")
    .attr("stroke", "#555")
    .attr("stroke-opacity", 0.4)
    .attr("stroke-width", 1.5)
  .selectAll("path")
    .data(root.links())
    .join("path")
      .attr("d", d3.linkHorizontal()
          .x(d => d.y)
          .y(d => d.x));
  
  const node = g.append("g")
      .attr("stroke-linejoin", "round")
      .attr("stroke-width", 3)
    .selectAll("g")
    .data(root.descendants())
    .join("g")
      .attr("transform", d => `translate(${d.y},${d.x})`);

  node.append("circle")
      .attr("fill", d => d.children ? "#555" : "#999")
      .attr("r", 2.5);

  node.append("text")
      .attr("dy", "0.31em")
      .attr("x", d => d.children ? -6 : 6)
      .attr("text-anchor", d => d.children ? "end" : "start")
      .text(d => d.data.name)
    .clone(true).lower()
      .attr("stroke", "white");
  
  return svg.node();
}
);
  main.variable(observer("data")).define("data", ["FileAttachment"], function(FileAttachment){return(
FileAttachment("flare-2.json").json()
)});
  main.variable(observer("tree")).define("tree", ["d3","width"], function(d3,width){return(
data => {
  const root = d3.hierarchy(data);
  root.dx = 10;
  root.dy = width / (root.height + 1);
  return d3.tree().nodeSize([root.dx, root.dy])(root);
}
)});
  main.variable(observer("width")).define("width", function(){return(
954
)});
  main.variable(observer("d3")).define("d3", ["require"], function(require){return(
require("d3@6")
)});
  return main;
}
