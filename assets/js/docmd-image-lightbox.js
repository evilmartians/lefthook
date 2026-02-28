/*!
 * docmd (v0.4.10)
 * Copyright (c) 2025-present docmd.io
 * License: MIT
 */
document.addEventListener("DOMContentLoaded",function(){const t=document.createElement("div");t.className="docmd-lightbox",t.innerHTML=`
    <div class="docmd-lightbox-content">
      <img src="" alt="">
      <div class="docmd-lightbox-caption"></div>
    </div>
    <div class="docmd-lightbox-close">&times;</div>
  `,document.body.appendChild(t);const l=t.querySelector("img"),d=t.querySelector(".docmd-lightbox-caption"),s=t.querySelector(".docmd-lightbox-close");document.querySelectorAll("img.lightbox, .image-gallery img").forEach(function(e){e.style.cursor="zoom-in",e.addEventListener("click",function(){const r=this.getAttribute("src");let i=this.getAttribute("alt")||"";const c=this.closest("figure");if(c){const n=c.querySelector("figcaption");n&&(i=n.textContent)}l.setAttribute("src",r),d.textContent=i,t.style.display="flex",document.body.style.overflow="hidden"})}),s.addEventListener("click",o),t.addEventListener("click",function(e){e.target===t&&o()}),document.addEventListener("keydown",function(e){e.key==="Escape"&&t.style.display==="flex"&&o()});function o(){t.style.display="none",document.body.style.overflow=""}});
