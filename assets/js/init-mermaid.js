import o from"https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs";(async function(){"use strict";let d=0;function s(){return document.documentElement.getAttribute("data-theme")==="dark"?"dark":"default"}async function t(){o.initialize({startOnLoad:!1,theme:s(),securityLevel:"loose"});let r=document.querySelectorAll('.mermaid:not([data-processed="true"])');for(let e of r){e.dataset.original||(e.dataset.original=e.textContent||"");let a=e.dataset.original;if(e.offsetParent!==null)try{let n=`mermaid-svg-${d++}`,{svg:i}=await o.render(n,a);e.innerHTML=i,e.setAttribute("data-processed","true")}catch{e.setAttribute("data-processed","error")}}}document.readyState==="loading"?document.addEventListener("DOMContentLoaded",t):t(),document.addEventListener("docmd:page-mounted",t),document.addEventListener("click",r=>{r.target?.closest(".docmd-tabs-nav-item, .collapsible-summary")&&setTimeout(t,50)}),new MutationObserver(r=>{for(let e of r)e.attributeName==="data-theme"&&(document.querySelectorAll(".mermaid").forEach(a=>a.removeAttribute("data-processed")),t())}).observe(document.documentElement,{attributes:!0,attributeFilter:["data-theme"]})})();
/**
 * --------------------------------------------------------------------
 * docmd : the minimalist, zero-config documentation generator.
 *
 * @package     @docmd/core (and ecosystem)
 * @website     https://docmd.io
 * @repository  https://github.com/docmd-io/docmd
 * @license     MIT
 * @copyright   Copyright (c) 2025-present docmd.io
 *
 * [docmd-source] - Please do not remove this header.
 * --------------------------------------------------------------------
 */
