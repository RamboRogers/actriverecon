// https://observablehq.com/@d3/radial-dendrogram@197
export default function define(runtime, observer) {
  const main = runtime.module();
  const fileAttachments = new Map([["flare-2.json",new URL("./files/e65374209781891f37dea1e7a6e1c5e020a3009b8aedf113b4c80942018887a1176ad4945cf14444603ff91d3da371b3b0d72419fa8d2ee0f6e815732475d5de",import.meta.url)]]);
  main.builtin("FileAttachment", runtime.fileAttachments(name => fileAttachments.get(name)));
  main.variable(observer()).define(["md"], function(md){return(
md`# Radial Dendrogram

D3’s [cluster layout](https://github.com/d3/d3-hierarchy/blob/master/README.md#cluster) produces node-link diagrams with leaf nodes at equal depth. These are less compact than [tidy trees](/@d3/radial-tidy-tree), but are useful for dendrograms, hierarchical clustering and [phylogenetic trees](/@mbostock/tree-of-life). See also the [Cartesian variant](/@d3/cluster-dendrogram).`
)});
  main.variable(observer("chart")).define("chart", ["tree","d3","data","autoBox"], function(tree,d3,data,autoBox)
{
  const root = tree(d3.hierarchy(data)
      .sort((a, b) => d3.ascending(a.data.name, b.data.name)));

  const svg = d3.create("svg");

  svg.append("g")
      .attr("fill", "none")
      .attr("stroke", "#555")
      .attr("stroke-opacity", 0.4)
      .attr("stroke-width", 1.5)
    .selectAll("path")
    .data(root.links())
    .join("path")
      .attr("d", d3.linkRadial()
          .angle(d => d.x)
          .radius(d => d.y));
  
  svg.append("g")
    .selectAll("circle")
    .data(root.descendants())
    .join("circle")
      .attr("transform", d => `
        rotate(${d.x * 180 / Math.PI - 90})
        translate(${d.y},0)
      `)
      .attr("fill", d => d.children ? "#555" : "#999")
      .attr("r", 2.5);

  svg.append("g")
      .attr("font-family", "sans-serif")
      .attr("font-size", 10)
      .attr("stroke-linejoin", "round")
      .attr("stroke-width", 3)
    .selectAll("text")
    .data(root.descendants())
    .join("text")
      .attr("transform", d => `
        rotate(${d.x * 180 / Math.PI - 90}) 
        translate(${d.y},0) 
        rotate(${d.x >= Math.PI ? 180 : 0})
      `)
      .attr("dy", "0.31em")
      .attr("x", d => d.x < Math.PI === !d.children ? 6 : -6)
      .attr("text-anchor", d => d.x < Math.PI === !d.children ? "start" : "end")
      .text(d => d.data.name)
    .clone(true).lower()
      .attr("stroke", "white");

  return svg.attr("viewBox", autoBox).node();
}
);
  main.variable(observer("autoBox")).define("autoBox", function(){return(
function autoBox() {
  document.body.appendChild(this);
  const {x, y, width, height} = this.getBBox();
  document.body.removeChild(this);
  return [x, y, width, height];
}
)});
  main.variable(observer("data")).define("data", ["FileAttachment"], function(FileAttachment){return(
FileAttachment("flare-2.json").json()
)});
  main.variable(observer("width")).define("width", function(){return(
975
)});
  main.variable(observer("radius")).define("radius", ["width"], function(width){return(
width / 2
)});
  main.variable(observer("tree")).define("tree", ["d3","radius"], function(d3,radius){return(
d3.cluster().size([2 * Math.PI, radius - 100])
)});
  main.variable(observer("d3")).define("d3", ["require"], function(require){return(
require("d3@6")
)});
  return main;
}
