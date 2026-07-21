var ut=globalThis,bt=ut.ShadowRoot&&(ut.ShadyCSS===void 0||ut.ShadyCSS.nativeShadow)&&"adoptedStyleSheets"in Document.prototype&&"replace"in CSSStyleSheet.prototype,Mt=Symbol(),It=new WeakMap,rt=class{constructor(t,e,i){if(this._$cssResult$=!0,i!==Mt)throw Error("CSSResult is not constructable. Use `unsafeCSS` or `css` instead.");this.cssText=t,this.t=e}get styleSheet(){let t=this.o,e=this.t;if(bt&&t===void 0){let i=e!==void 0&&e.length===1;i&&(t=It.get(e)),t===void 0&&((this.o=t=new CSSStyleSheet).replaceSync(this.cssText),i&&It.set(e,t))}return t}toString(){return this.cssText}},Ft=r=>new rt(typeof r=="string"?r:r+"",void 0,Mt),u=(r,...t)=>{let e=r.length===1?r[0]:t.reduce((i,s,o)=>i+(a=>{if(a._$cssResult$===!0)return a.cssText;if(typeof a=="number")return a;throw Error("Value passed to 'css' function must be a 'css' function result: "+a+". Use 'unsafeCSS' to pass non-literal values, but take care to ensure page security.")})(s)+r[o+1],r[0]);return new rt(e,r,Mt)},Jt=(r,t)=>{if(bt)r.adoptedStyleSheets=t.map(e=>e instanceof CSSStyleSheet?e:e.styleSheet);else for(let e of t){let i=document.createElement("style"),s=ut.litNonce;s!==void 0&&i.setAttribute("nonce",s),i.textContent=e.cssText,r.appendChild(i)}},qt=bt?r=>r:r=>r instanceof CSSStyleSheet?(t=>{let e="";for(let i of t.cssRules)e+=i.cssText;return Ft(e)})(r):r;var{is:ye,defineProperty:ge,getOwnPropertyDescriptor:$e,getOwnPropertyNames:xe,getOwnPropertySymbols:_e,getPrototypeOf:Ae}=Object,ft=globalThis,Wt=ft.trustedTypes,ke=Wt?Wt.emptyScript:"",we=ft.reactiveElementPolyfillSupport,it=(r,t)=>r,ot={toAttribute(r,t){switch(t){case Boolean:r=r?ke:null;break;case Object:case Array:r=r==null?r:JSON.stringify(r)}return r},fromAttribute(r,t){let e=r;switch(t){case Boolean:e=r!==null;break;case Number:e=r===null?null:Number(r);break;case Object:case Array:try{e=JSON.parse(r)}catch{e=null}}return e}},vt=(r,t)=>!ye(r,t),Kt={attribute:!0,type:String,converter:ot,reflect:!1,useDefault:!1,hasChanged:vt};Symbol.metadata??=Symbol("metadata"),ft.litPropertyMetadata??=new WeakMap;var R=class extends HTMLElement{static addInitializer(t){this._$Ei(),(this.l??=[]).push(t)}static get observedAttributes(){return this.finalize(),this._$Eh&&[...this._$Eh.keys()]}static createProperty(t,e=Kt){if(e.state&&(e.attribute=!1),this._$Ei(),this.prototype.hasOwnProperty(t)&&((e=Object.create(e)).wrapped=!0),this.elementProperties.set(t,e),!e.noAccessor){let i=Symbol(),s=this.getPropertyDescriptor(t,i,e);s!==void 0&&ge(this.prototype,t,s)}}static getPropertyDescriptor(t,e,i){let{get:s,set:o}=$e(this.prototype,t)??{get(){return this[e]},set(a){this[e]=a}};return{get:s,set(a){let p=s?.call(this);o?.call(this,a),this.requestUpdate(t,p,i)},configurable:!0,enumerable:!0}}static getPropertyOptions(t){return this.elementProperties.get(t)??Kt}static _$Ei(){if(this.hasOwnProperty(it("elementProperties")))return;let t=Ae(this);t.finalize(),t.l!==void 0&&(this.l=[...t.l]),this.elementProperties=new Map(t.elementProperties)}static finalize(){if(this.hasOwnProperty(it("finalized")))return;if(this.finalized=!0,this._$Ei(),this.hasOwnProperty(it("properties"))){let e=this.properties,i=[...xe(e),..._e(e)];for(let s of i)this.createProperty(s,e[s])}let t=this[Symbol.metadata];if(t!==null){let e=litPropertyMetadata.get(t);if(e!==void 0)for(let[i,s]of e)this.elementProperties.set(i,s)}this._$Eh=new Map;for(let[e,i]of this.elementProperties){let s=this._$Eu(e,i);s!==void 0&&this._$Eh.set(s,e)}this.elementStyles=this.finalizeStyles(this.styles)}static finalizeStyles(t){let e=[];if(Array.isArray(t)){let i=new Set(t.flat(1/0).reverse());for(let s of i)e.unshift(qt(s))}else t!==void 0&&e.push(qt(t));return e}static _$Eu(t,e){let i=e.attribute;return i===!1?void 0:typeof i=="string"?i:typeof t=="string"?t.toLowerCase():void 0}constructor(){super(),this._$Ep=void 0,this.isUpdatePending=!1,this.hasUpdated=!1,this._$Em=null,this._$Ev()}_$Ev(){this._$ES=new Promise(t=>this.enableUpdating=t),this._$AL=new Map,this._$E_(),this.requestUpdate(),this.constructor.l?.forEach(t=>t(this))}addController(t){(this._$EO??=new Set).add(t),this.renderRoot!==void 0&&this.isConnected&&t.hostConnected?.()}removeController(t){this._$EO?.delete(t)}_$E_(){let t=new Map,e=this.constructor.elementProperties;for(let i of e.keys())this.hasOwnProperty(i)&&(t.set(i,this[i]),delete this[i]);t.size>0&&(this._$Ep=t)}createRenderRoot(){let t=this.shadowRoot??this.attachShadow(this.constructor.shadowRootOptions);return Jt(t,this.constructor.elementStyles),t}connectedCallback(){this.renderRoot??=this.createRenderRoot(),this.enableUpdating(!0),this._$EO?.forEach(t=>t.hostConnected?.())}enableUpdating(t){}disconnectedCallback(){this._$EO?.forEach(t=>t.hostDisconnected?.())}attributeChangedCallback(t,e,i){this._$AK(t,i)}_$ET(t,e){let i=this.constructor.elementProperties.get(t),s=this.constructor._$Eu(t,i);if(s!==void 0&&i.reflect===!0){let o=(i.converter?.toAttribute!==void 0?i.converter:ot).toAttribute(e,i.type);this._$Em=t,o==null?this.removeAttribute(s):this.setAttribute(s,o),this._$Em=null}}_$AK(t,e){let i=this.constructor,s=i._$Eh.get(t);if(s!==void 0&&this._$Em!==s){let o=i.getPropertyOptions(s),a=typeof o.converter=="function"?{fromAttribute:o.converter}:o.converter?.fromAttribute!==void 0?o.converter:ot;this._$Em=s;let p=a.fromAttribute(e,o.type);this[s]=p??this._$Ej?.get(s)??p,this._$Em=null}}requestUpdate(t,e,i,s=!1,o){if(t!==void 0){let a=this.constructor;if(s===!1&&(o=this[t]),i??=a.getPropertyOptions(t),!((i.hasChanged??vt)(o,e)||i.useDefault&&i.reflect&&o===this._$Ej?.get(t)&&!this.hasAttribute(a._$Eu(t,i))))return;this.C(t,e,i)}this.isUpdatePending===!1&&(this._$ES=this._$EP())}C(t,e,{useDefault:i,reflect:s,wrapped:o},a){i&&!(this._$Ej??=new Map).has(t)&&(this._$Ej.set(t,a??e??this[t]),o!==!0||a!==void 0)||(this._$AL.has(t)||(this.hasUpdated||i||(e=void 0),this._$AL.set(t,e)),s===!0&&this._$Em!==t&&(this._$Eq??=new Set).add(t))}async _$EP(){this.isUpdatePending=!0;try{await this._$ES}catch(e){Promise.reject(e)}let t=this.scheduleUpdate();return t!=null&&await t,!this.isUpdatePending}scheduleUpdate(){return this.performUpdate()}performUpdate(){if(!this.isUpdatePending)return;if(!this.hasUpdated){if(this.renderRoot??=this.createRenderRoot(),this._$Ep){for(let[s,o]of this._$Ep)this[s]=o;this._$Ep=void 0}let i=this.constructor.elementProperties;if(i.size>0)for(let[s,o]of i){let{wrapped:a}=o,p=this[s];a!==!0||this._$AL.has(s)||p===void 0||this.C(s,void 0,o,p)}}let t=!1,e=this._$AL;try{t=this.shouldUpdate(e),t?(this.willUpdate(e),this._$EO?.forEach(i=>i.hostUpdate?.()),this.update(e)):this._$EM()}catch(i){throw t=!1,this._$EM(),i}t&&this._$AE(e)}willUpdate(t){}_$AE(t){this._$EO?.forEach(e=>e.hostUpdated?.()),this.hasUpdated||(this.hasUpdated=!0,this.firstUpdated(t)),this.updated(t)}_$EM(){this._$AL=new Map,this.isUpdatePending=!1}get updateComplete(){return this.getUpdateComplete()}getUpdateComplete(){return this._$ES}shouldUpdate(t){return!0}update(t){this._$Eq&&=this._$Eq.forEach(e=>this._$ET(e,this[e])),this._$EM()}updated(t){}firstUpdated(t){}};R.elementStyles=[],R.shadowRootOptions={mode:"open"},R[it("elementProperties")]=new Map,R[it("finalized")]=new Map,we?.({ReactiveElement:R}),(ft.reactiveElementVersions??=[]).push("2.1.2");var Rt=globalThis,Gt=r=>r,yt=Rt.trustedTypes,Qt=yt?yt.createPolicy("lit-html",{createHTML:r=>r}):void 0,Tt="$lit$",T=`lit$${Math.random().toFixed(9).slice(2)}$`,Nt="?"+T,Ee=`<${Nt}>`,K=document,nt=()=>K.createComment(""),lt=r=>r===null||typeof r!="object"&&typeof r!="function",Bt=Array.isArray,se=r=>Bt(r)||typeof r?.[Symbol.iterator]=="function",Dt=`[ 	
\f\r]`,at=/<(?:(!--|\/[^a-zA-Z])|(\/?[a-zA-Z][^>\s]*)|(\/?$))/g,Yt=/-->/g,Zt=/>/g,J=RegExp(`>|${Dt}(?:([^\\s"'>=/]+)(${Dt}*=${Dt}*(?:[^ 	
\f\r"'\`<>=]|("|')|))|$)`,"g"),Xt=/'/g,te=/"/g,re=/^(?:script|style|textarea|title)$/i,jt=r=>(t,...e)=>({_$litType$:r,strings:t,values:e}),d=jt(1),os=jt(2),as=jt(3),N=Symbol.for("lit-noChange"),l=Symbol.for("lit-nothing"),ee=new WeakMap,W=K.createTreeWalker(K,129);function ie(r,t){if(!Bt(r)||!r.hasOwnProperty("raw"))throw Error("invalid template strings array");return Qt!==void 0?Qt.createHTML(t):t}var oe=(r,t)=>{let e=r.length-1,i=[],s,o=t===2?"<svg>":t===3?"<math>":"",a=at;for(let p=0;p<e;p++){let h=r[p],y,x,c=-1,g=0;for(;g<h.length&&(a.lastIndex=g,x=a.exec(h),x!==null);)g=a.lastIndex,a===at?x[1]==="!--"?a=Yt:x[1]!==void 0?a=Zt:x[2]!==void 0?(re.test(x[2])&&(s=RegExp("</"+x[2],"g")),a=J):x[3]!==void 0&&(a=J):a===J?x[0]===">"?(a=s??at,c=-1):x[1]===void 0?c=-2:(c=a.lastIndex-x[2].length,y=x[1],a=x[3]===void 0?J:x[3]==='"'?te:Xt):a===te||a===Xt?a=J:a===Yt||a===Zt?a=at:(a=J,s=void 0);let v=a===J&&r[p+1].startsWith("/>")?" ":"";o+=a===at?h+Ee:c>=0?(i.push(y),h.slice(0,c)+Tt+h.slice(c)+T+v):h+T+(c===-2?p:v)}return[ie(r,o+(r[e]||"<?>")+(t===2?"</svg>":t===3?"</math>":"")),i]},ht=class r{constructor({strings:t,_$litType$:e},i){let s;this.parts=[];let o=0,a=0,p=t.length-1,h=this.parts,[y,x]=oe(t,e);if(this.el=r.createElement(y,i),W.currentNode=this.el.content,e===2||e===3){let c=this.el.content.firstChild;c.replaceWith(...c.childNodes)}for(;(s=W.nextNode())!==null&&h.length<p;){if(s.nodeType===1){if(s.hasAttributes())for(let c of s.getAttributeNames())if(c.endsWith(Tt)){let g=x[a++],v=s.getAttribute(c).split(T),_=/([.?@])?(.*)/.exec(g);h.push({type:1,index:o,name:_[2],strings:v,ctor:_[1]==="."?$t:_[1]==="?"?xt:_[1]==="@"?_t:Q}),s.removeAttribute(c)}else c.startsWith(T)&&(h.push({type:6,index:o}),s.removeAttribute(c));if(re.test(s.tagName)){let c=s.textContent.split(T),g=c.length-1;if(g>0){s.textContent=yt?yt.emptyScript:"";for(let v=0;v<g;v++)s.append(c[v],nt()),W.nextNode(),h.push({type:2,index:++o});s.append(c[g],nt())}}}else if(s.nodeType===8)if(s.data===Nt)h.push({type:2,index:o});else{let c=-1;for(;(c=s.data.indexOf(T,c+1))!==-1;)h.push({type:7,index:o}),c+=T.length-1}o++}}static createElement(t,e){let i=K.createElement("template");return i.innerHTML=t,i}};function G(r,t,e=r,i){if(t===N)return t;let s=i!==void 0?e._$Co?.[i]:e._$Cl,o=lt(t)?void 0:t._$litDirective$;return s?.constructor!==o&&(s?._$AO?.(!1),o===void 0?s=void 0:(s=new o(r),s._$AT(r,e,i)),i!==void 0?(e._$Co??=[])[i]=s:e._$Cl=s),s!==void 0&&(t=G(r,s._$AS(r,t.values),s,i)),t}var gt=class{constructor(t,e){this._$AV=[],this._$AN=void 0,this._$AD=t,this._$AM=e}get parentNode(){return this._$AM.parentNode}get _$AU(){return this._$AM._$AU}u(t){let{el:{content:e},parts:i}=this._$AD,s=(t?.creationScope??K).importNode(e,!0);W.currentNode=s;let o=W.nextNode(),a=0,p=0,h=i[0];for(;h!==void 0;){if(a===h.index){let y;h.type===2?y=new X(o,o.nextSibling,this,t):h.type===1?y=new h.ctor(o,h.name,h.strings,this,t):h.type===6&&(y=new At(o,this,t)),this._$AV.push(y),h=i[++p]}a!==h?.index&&(o=W.nextNode(),a++)}return W.currentNode=K,s}p(t){let e=0;for(let i of this._$AV)i!==void 0&&(i.strings!==void 0?(i._$AI(t,i,e),e+=i.strings.length-2):i._$AI(t[e])),e++}},X=class r{get _$AU(){return this._$AM?._$AU??this._$Cv}constructor(t,e,i,s){this.type=2,this._$AH=l,this._$AN=void 0,this._$AA=t,this._$AB=e,this._$AM=i,this.options=s,this._$Cv=s?.isConnected??!0}get parentNode(){let t=this._$AA.parentNode,e=this._$AM;return e!==void 0&&t?.nodeType===11&&(t=e.parentNode),t}get startNode(){return this._$AA}get endNode(){return this._$AB}_$AI(t,e=this){t=G(this,t,e),lt(t)?t===l||t==null||t===""?(this._$AH!==l&&this._$AR(),this._$AH=l):t!==this._$AH&&t!==N&&this._(t):t._$litType$!==void 0?this.$(t):t.nodeType!==void 0?this.T(t):se(t)?this.k(t):this._(t)}O(t){return this._$AA.parentNode.insertBefore(t,this._$AB)}T(t){this._$AH!==t&&(this._$AR(),this._$AH=this.O(t))}_(t){this._$AH!==l&&lt(this._$AH)?this._$AA.nextSibling.data=t:this.T(K.createTextNode(t)),this._$AH=t}$(t){let{values:e,_$litType$:i}=t,s=typeof i=="number"?this._$AC(t):(i.el===void 0&&(i.el=ht.createElement(ie(i.h,i.h[0]),this.options)),i);if(this._$AH?._$AD===s)this._$AH.p(e);else{let o=new gt(s,this),a=o.u(this.options);o.p(e),this.T(a),this._$AH=o}}_$AC(t){let e=ee.get(t.strings);return e===void 0&&ee.set(t.strings,e=new ht(t)),e}k(t){Bt(this._$AH)||(this._$AH=[],this._$AR());let e=this._$AH,i,s=0;for(let o of t)s===e.length?e.push(i=new r(this.O(nt()),this.O(nt()),this,this.options)):i=e[s],i._$AI(o),s++;s<e.length&&(this._$AR(i&&i._$AB.nextSibling,s),e.length=s)}_$AR(t=this._$AA.nextSibling,e){for(this._$AP?.(!1,!0,e);t!==this._$AB;){let i=Gt(t).nextSibling;Gt(t).remove(),t=i}}setConnected(t){this._$AM===void 0&&(this._$Cv=t,this._$AP?.(t))}},Q=class{get tagName(){return this.element.tagName}get _$AU(){return this._$AM._$AU}constructor(t,e,i,s,o){this.type=1,this._$AH=l,this._$AN=void 0,this.element=t,this.name=e,this._$AM=s,this.options=o,i.length>2||i[0]!==""||i[1]!==""?(this._$AH=Array(i.length-1).fill(new String),this.strings=i):this._$AH=l}_$AI(t,e=this,i,s){let o=this.strings,a=!1;if(o===void 0)t=G(this,t,e,0),a=!lt(t)||t!==this._$AH&&t!==N,a&&(this._$AH=t);else{let p=t,h,y;for(t=o[0],h=0;h<o.length-1;h++)y=G(this,p[i+h],e,h),y===N&&(y=this._$AH[h]),a||=!lt(y)||y!==this._$AH[h],y===l?t=l:t!==l&&(t+=(y??"")+o[h+1]),this._$AH[h]=y}a&&!s&&this.j(t)}j(t){t===l?this.element.removeAttribute(this.name):this.element.setAttribute(this.name,t??"")}},$t=class extends Q{constructor(){super(...arguments),this.type=3}j(t){this.element[this.name]=t===l?void 0:t}},xt=class extends Q{constructor(){super(...arguments),this.type=4}j(t){this.element.toggleAttribute(this.name,!!t&&t!==l)}},_t=class extends Q{constructor(t,e,i,s,o){super(t,e,i,s,o),this.type=5}_$AI(t,e=this){if((t=G(this,t,e,0)??l)===N)return;let i=this._$AH,s=t===l&&i!==l||t.capture!==i.capture||t.once!==i.once||t.passive!==i.passive,o=t!==l&&(i===l||s);s&&this.element.removeEventListener(this.name,this,i),o&&this.element.addEventListener(this.name,this,t),this._$AH=t}handleEvent(t){typeof this._$AH=="function"?this._$AH.call(this.options?.host??this.element,t):this._$AH.handleEvent(t)}},At=class{constructor(t,e,i){this.element=t,this.type=6,this._$AN=void 0,this._$AM=e,this.options=i}get _$AU(){return this._$AM._$AU}_$AI(t){G(this,t)}},ae={M:Tt,P:T,A:Nt,C:1,L:oe,R:gt,D:se,V:G,I:X,H:Q,N:xt,U:_t,B:$t,F:At},Se=Rt.litHtmlPolyfillSupport;Se?.(ht,X),(Rt.litHtmlVersions??=[]).push("3.3.3");var ne=(r,t,e)=>{let i=e?.renderBefore??t,s=i._$litPart$;if(s===void 0){let o=e?.renderBefore??null;i._$litPart$=s=new X(t.insertBefore(nt(),o),o,void 0,e??{})}return s._$AI(r),s};var Ht=globalThis,m=class extends R{constructor(){super(...arguments),this.renderOptions={host:this},this._$Do=void 0}createRenderRoot(){let t=super.createRenderRoot();return this.renderOptions.renderBefore??=t.firstChild,t}update(t){let e=this.render();this.hasUpdated||(this.renderOptions.isConnected=this.isConnected),super.update(t),this._$Do=ne(e,this.renderRoot,this.renderOptions)}connectedCallback(){super.connectedCallback(),this._$Do?.setConnected(!0)}disconnectedCallback(){super.disconnectedCallback(),this._$Do?.setConnected(!1)}render(){return N}};m._$litElement$=!0,m.finalized=!0,Ht.litElementHydrateSupport?.({LitElement:m});var Ce=Ht.litElementPolyfillSupport;Ce?.({LitElement:m});(Ht.litElementVersions??=[]).push("4.2.2");var ze={attribute:!0,type:String,converter:ot,reflect:!1,hasChanged:vt},Pe=(r=ze,t,e)=>{let{kind:i,metadata:s}=e,o=globalThis.litPropertyMetadata.get(s);if(o===void 0&&globalThis.litPropertyMetadata.set(s,o=new Map),i==="setter"&&((r=Object.create(r)).wrapped=!0),o.set(e.name,r),i==="accessor"){let{name:a}=e;return{set(p){let h=t.get.call(this);t.set.call(this,p),this.requestUpdate(a,h,r,!0,p)},init(p){return p!==void 0&&this.C(a,void 0,r,p),p}}}if(i==="setter"){let{name:a}=e;return function(p){let h=this[a];t.call(this,p),this.requestUpdate(a,h,r,!0,p)}}throw Error("Unsupported decorator location: "+i)};function n(r){return(t,e)=>typeof e=="object"?Pe(r,t,e):((i,s,o)=>{let a=s.hasOwnProperty(o);return s.constructor.createProperty(o,i),a?Object.getOwnPropertyDescriptor(s,o):void 0})(r,t,e)}function dt(r){return n({...r,state:!0,attribute:!1})}function k(r,t,e){r.setFormValue(t,t)}function M(r,t,e="",i){r.setValidity(t,e,i)}function q(r){r.setValidity({})}function j(r,t,e="Please fill out this field."){return r?{flags:{customError:!0},message:r}:t?{flags:{valueMissing:!0},message:e}:{flags:{},message:""}}function b(r,t){customElements.get(r)||customElements.define(r,t)}var f=u`
  :host {
    box-sizing: border-box;
    font-family: var(--mb-font-body);
    color: var(--mb-color-fg);
    max-inline-size: 100%;
    overflow-wrap: anywhere;
  }

  :host *,
  :host *::before,
  :host *::after {
    box-sizing: border-box;
  }

  :host([hidden]) {
    display: none !important;
  }

  .control:focus-visible,
  button:focus-visible,
  a:focus-visible,
  select:focus-visible,
  textarea:focus-visible,
  input:focus-visible {
    outline: var(--mb-focus-ring);
    outline-offset: var(--mb-focus-offset);
  }

  @media (prefers-reduced-motion: reduce) {
    :host,
    :host * {
      transition: none !important;
      animation: none !important;
    }
  }
`,tt=u`
  .field {
    display: flex;
    flex-direction: column;
    gap: var(--mb-space-1);
    inline-size: 100%;
  }

  .label {
    font-size: var(--mb-font-size-sm);
    font-weight: 600;
    color: var(--mb-color-fg);
  }

  .label.visually-hidden {
    position: absolute;
    inline-size: 1px;
    block-size: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }

  .hint,
  .error {
    font-size: var(--mb-font-size-sm);
    margin: 0;
  }

  .hint {
    color: var(--mb-color-muted);
  }

  .error {
    color: var(--mb-color-danger);
  }

  .control {
    inline-size: 100%;
    max-inline-size: 100%;
    min-block-size: 2.5rem;
    min-inline-size: 0;
    padding-block: var(--mb-space-2);
    padding-inline: var(--mb-space-3);
    border: 1px solid var(--mb-color-border);
    border-radius: var(--mb-radius-md);
    background: var(--mb-color-surface);
    color: var(--mb-color-fg);
    font: inherit;
    transition:
      border-color var(--mb-transition),
      box-shadow var(--mb-transition);
  }

  .control:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }

  :host([invalid]) .control {
    border-color: var(--mb-color-danger);
  }

  :host([density='compact']) .field {
    gap: 0;
  }

  :host([density='compact']) .control {
    min-block-size: 2.1rem;
    padding-block: 0.2rem;
    padding-inline: var(--mb-space-2);
    font-size: var(--mb-font-size-sm);
  }

  :host([density='compact']) textarea.control {
    min-block-size: 2.1rem;
  }
`;function et(r,t,e){return r?{labelText:r,hideVisually:t,controlAriaLabel:""}:{labelText:"",hideVisually:!1,controlAriaLabel:e}}var Oe=Object.defineProperty,O=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Oe(t,e,s),s},S=class extends m{constructor(){super(...arguments),this.variant="primary",this.size="md",this.type="button",this.disabled=!1,this.loading=!1,this.name="",this.value="",this.href="",this.target="",this.rel="",this.iconOnly=!1,this.#t=this.attachInternals(),this.#e=!1}static{this.formAssociated=!0}static{this.styles=[f,u`
      :host {
        display: inline-block;
      }

      .base {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        gap: var(--mb-space-2);
        max-inline-size: 100%;
        border: 1px solid transparent;
        border-radius: var(--mb-radius-md);
        font: inherit;
        font-weight: 600;
        cursor: pointer;
        white-space: normal;
        text-align: center;
        text-decoration: none;
        overflow-wrap: anywhere;
        transition:
          background-color var(--mb-transition),
          color var(--mb-transition),
          border-color var(--mb-transition),
          opacity var(--mb-transition);
      }

      .base:disabled,
      .base[aria-disabled='true'] {
        cursor: not-allowed;
        opacity: 0.55;
        pointer-events: none;
      }

      :host([size='sm']) .base {
        min-block-size: 2rem;
        padding-inline: var(--mb-space-3);
        font-size: var(--mb-font-size-sm);
      }

      :host([size='md']) .base {
        min-block-size: 2.5rem;
        padding-inline: var(--mb-space-4);
        font-size: var(--mb-font-size-md);
      }

      :host([size='lg']) .base {
        min-block-size: 3rem;
        padding-inline: var(--mb-space-5);
        font-size: var(--mb-font-size-lg);
      }

      :host([icon-only][size='sm']) .base {
        min-inline-size: 2rem;
        padding-inline: 0;
      }

      :host([icon-only][size='md']) .base,
      :host([icon-only]:not([size])) .base {
        min-inline-size: 2.5rem;
        padding-inline: 0;
      }

      :host([icon-only][size='lg']) .base {
        min-inline-size: 3rem;
        padding-inline: 0;
      }

      :host([variant='primary']) .base {
        background: var(--mb-color-accent);
        color: var(--mb-color-on-accent);
      }

      :host([variant='secondary']) .base {
        background: var(--mb-color-surface);
        color: var(--mb-color-fg);
        border-color: var(--mb-color-border);
      }

      :host([variant='ghost']) .base {
        background: transparent;
        color: var(--mb-color-accent);
      }

      :host([variant='danger']) .base {
        background: var(--mb-color-danger);
        color: var(--mb-color-on-danger);
      }

      .spinner {
        inline-size: 1em;
        block-size: 1em;
        border: 2px solid currentColor;
        border-inline-end-color: transparent;
        border-radius: 50%;
        animation: spin 0.7s linear infinite;
      }

      @keyframes spin {
        to {
          transform: rotate(360deg);
        }
      }
    `]}#t;#e;get#s(){return this.disabled||this.loading||this.#e}get#r(){return!!this.href}get#o(){return this.getAttribute("aria-label")??""}formDisabledCallback(t){this.#e=t,this.requestUpdate()}#i(t){if(this.#s){t.preventDefault(),t.stopImmediatePropagation();return}if(this.#r)return;let e=this.#t.form;e&&(this.type==="submit"?(this.name&&k(this.#t,this.value),e.requestSubmit(),queueMicrotask(()=>k(this.#t,null))):this.type==="reset"&&e.reset())}render(){let t=d`
      ${this.loading?d`<span class="spinner" aria-hidden="true"></span>`:l}
      <slot></slot>
    `,e=this.#o||l;return this.#r?d`
        <a
          part="base"
          class="base"
          href=${this.#s?l:this.href}
          target=${this.target||l}
          rel=${this.rel||(this.target==="_blank"?"noopener noreferrer":l)}
          aria-disabled=${this.#s?"true":"false"}
          aria-busy=${this.loading?"true":"false"}
          aria-label=${e}
          @click=${this.#i}
        >
          ${t}
        </a>
      `:d`
      <button
        part="base"
        class="base"
        type="button"
        ?disabled=${this.#s}
        aria-busy=${this.loading?"true":"false"}
        aria-label=${e}
        @click=${this.#i}
      >
        ${t}
      </button>
    `}};O([n({reflect:!0})],S.prototype,"variant");O([n({reflect:!0})],S.prototype,"size");O([n({reflect:!0})],S.prototype,"type");O([n({type:Boolean,reflect:!0})],S.prototype,"disabled");O([n({type:Boolean,reflect:!0})],S.prototype,"loading");O([n({reflect:!0})],S.prototype,"name");O([n()],S.prototype,"value");O([n({reflect:!0})],S.prototype,"href");O([n({reflect:!0})],S.prototype,"target");O([n({reflect:!0})],S.prototype,"rel");O([n({type:Boolean,reflect:!0,attribute:"icon-only"})],S.prototype,"iconOnly");b("mb-button",S);var Le=Object.defineProperty,Ue=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Le(t,e,s),s},wt=class extends m{constructor(){super(...arguments),this.variant="neutral"}static{this.styles=[f,u`
      :host {
        display: inline-flex;
      }

      span {
        display: inline-flex;
        align-items: center;
        gap: var(--mb-space-1);
        padding-block: 0.15rem;
        padding-inline: var(--mb-space-2);
        border-radius: var(--mb-radius-sm);
        font-size: var(--mb-font-size-sm);
        font-weight: 600;
        line-height: 1.3;
        background: var(--mb-color-border);
        color: var(--mb-color-fg);
      }

      :host([variant='success']) span {
        background: var(--mb-color-success-soft);
        color: var(--mb-color-success);
      }

      :host([variant='warning']) span {
        background: var(--mb-color-warning-soft);
        color: var(--mb-color-warning);
      }

      :host([variant='danger']) span {
        background: var(--mb-color-danger-soft);
        color: var(--mb-color-danger);
      }

      :host([variant='info']) span {
        background: var(--mb-color-info-soft);
        color: var(--mb-color-info);
      }
    `]}render(){return d`<span part="base"><slot></slot></span>`}};Ue([n({reflect:!0})],wt.prototype,"variant");b("mb-badge",wt);var Me=Object.defineProperty,qe=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Me(t,e,s),s},Et=class extends m{constructor(){super(...arguments),this.variant="info"}static{this.styles=[f,u`
      :host {
        display: block;
        inline-size: 100%;
      }

      .alert {
        padding-block: var(--mb-space-3);
        padding-inline: var(--mb-space-4);
        border-radius: var(--mb-radius-md);
        border-inline-start: 4px solid currentColor;
        background: var(--mb-color-info-soft);
        color: var(--mb-color-info);
        overflow-wrap: anywhere;
        max-inline-size: 100%;
      }

      :host([variant='success']) .alert {
        background: var(--mb-color-success-soft);
        color: var(--mb-color-success);
      }

      :host([variant='warning']) .alert {
        background: var(--mb-color-warning-soft);
        color: var(--mb-color-warning);
      }

      :host([variant='danger']) .alert {
        background: var(--mb-color-danger-soft);
        color: var(--mb-color-danger);
      }
    `]}get#t(){return this.variant==="warning"||this.variant==="danger"?"alert":"status"}render(){return d`
      <div part="base" class="alert" role=${this.#t}>
        <slot></slot>
      </div>
    `}};qe([n({reflect:!0})],Et.prototype,"variant");b("mb-alert",Et);var De=Object.defineProperty,le=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&De(t,e,s),s},ct=class extends m{constructor(){super(...arguments),this._hasHeader=!1,this._hasFooter=!1}static{this.styles=[f,u`
      :host {
        display: block;
        inline-size: 100%;
      }

      .card {
        background: var(--mb-color-surface);
        border: 1px solid var(--mb-color-border);
        border-radius: var(--mb-radius-lg);
        overflow: clip;
        max-inline-size: 100%;
      }

      .header,
      .body,
      .footer {
        padding-block: var(--mb-space-4);
        padding-inline: var(--mb-space-5);
        min-inline-size: 0;
        overflow-wrap: anywhere;
      }

      .header {
        display: none;
        border-block-end: 1px solid var(--mb-color-border);
        font-family: var(--mb-font-display);
        font-weight: 650;
      }

      .footer {
        display: none;
        border-block-start: 1px solid var(--mb-color-border);
      }

      :host([data-has-header]) .header,
      :host([data-has-footer]) .footer {
        display: block;
      }

      ::slotted([slot='header']),
      ::slotted([slot='footer']) {
        display: block;
      }
    `]}#t(t){let e=t.target;this._hasHeader=e.assignedNodes({flatten:!0}).length>0,this.toggleAttribute("data-has-header",this._hasHeader)}#e(t){let e=t.target;this._hasFooter=e.assignedNodes({flatten:!0}).length>0,this.toggleAttribute("data-has-footer",this._hasFooter)}render(){return d`
      <article part="card" class="card">
        <header class="header" part="header">
          <slot name="header" @slotchange=${this.#t}></slot>
        </header>
        <div class="body" part="body">
          <slot></slot>
        </div>
        <footer class="footer" part="footer">
          <slot name="footer" @slotchange=${this.#e}></slot>
        </footer>
      </article>
    `}};le([dt()],ct.prototype,"_hasHeader");le([dt()],ct.prototype,"_hasFooter");b("mb-card",ct);var Re=Object.defineProperty,A=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Re(t,e,s),s},$=class extends m{constructor(){super(...arguments),this.label="",this.hint="",this.error="",this.value="",this.name="",this.placeholder="",this.type="text",this.disabled=!1,this.required=!1,this.invalid=!1,this.density="default",this.hideLabel=!1,this.min="",this.max="",this.step="",this.accept="",this.multiple=!1,this.#t=this.attachInternals(),this.#e=!1,this.#r="",this.#o=!1,this.#i=!1}static{this.formAssociated=!0}static{this.styles=[f,tt,u`
      :host {
        display: block;
      }

      input[type='file'].control {
        padding-block: var(--mb-space-2);
      }
    `]}#t;#e;#s;#r;#o;#i;get#d(){return this.disabled||this.#e}get#a(){return this.type==="file"}get#n(){return this.getAttribute("aria-label")??""}connectedCallback(){super.connectedCallback(),this.#o||(this.#r=this.value,this.#o=!0)}firstUpdated(){this.#s=this.renderRoot.querySelector("input")??void 0,this.#h()}updated(t){(t.has("value")||t.has("required")||t.has("error")||t.has("disabled")||t.has("name")||t.has("type"))&&this.#h()}formDisabledCallback(t){this.#e=t,this.requestUpdate()}formResetCallback(){this.#i=!1,this.value=this.#r,this.error="",this.invalid=!1,this.#a&&this.#s&&(this.#s.value="")}#l(){let t=this.#s?.files;if(!this.name||!t?.length){k(this.#t,null);return}if(t.length===1){k(this.#t,t[0]);return}let e=new FormData;for(let i of t)e.append(this.name,i);k(this.#t,e)}#h(){this.#a?this.#l():k(this.#t,this.name?this.value:null);let t=this.required&&(this.#a?!this.#s?.files?.length:!this.value),{flags:e,message:i}=j(this.error,t);i?(M(this.#t,e,i,this.#s),this.invalid=!!this.error||this.#i):(q(this.#t),this.invalid=!1)}#c(t){let e=t.target;this.#i=!0,this.#a||(this.value=e.value),this.#h(),this.dispatchEvent(new CustomEvent("mb-input",{detail:{value:this.value,files:e.files},bubbles:!0,composed:!0}))}#p(t){let e=t.target;this.#i=!0,this.#a||(this.value=e.value),this.#h(),this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value,files:e.files},bubbles:!0,composed:!0}))}#m(t){if(t.key!=="Enter"||t.defaultPrevented||this.#a)return;let e=this.#t.form;e&&(t.preventDefault(),e.requestSubmit())}render(){let t=[this.hint&&!this.error?"hint":"",this.error?"error":""].filter(Boolean).join(" "),{labelText:e,hideVisually:i,controlAriaLabel:s}=et(this.label,this.hideLabel,this.#n);return d`
      <div class="field">
        ${e?d`<label
              part="label"
              class="label${i?" visually-hidden":""}"
              for="control"
              >${e}</label
            >`:l}
        <input
          id="control"
          part="control"
          class="control"
          .type=${this.type}
          .value=${this.#a?"":this.value}
          name=${this.name||l}
          placeholder=${this.placeholder||l}
          min=${this.type==="number"&&this.min!==""?this.min:l}
          max=${this.type==="number"&&this.max!==""?this.max:l}
          step=${this.type==="number"&&this.step!==""?this.step:l}
          accept=${this.#a&&this.accept?this.accept:l}
          ?multiple=${this.#a&&this.multiple}
          ?disabled=${this.#d}
          ?required=${this.required}
          aria-invalid=${this.invalid?"true":"false"}
          aria-label=${s||l}
          aria-describedby=${t||l}
          @input=${this.#c}
          @change=${this.#p}
          @keydown=${this.#m}
        />
        ${this.hint&&!this.error?d`<p id="hint" class="hint">${this.hint}</p>`:l}
        ${this.error?d`<p id="error" class="error" role="alert">${this.error}</p>`:l}
      </div>
    `}};A([n()],$.prototype,"label");A([n()],$.prototype,"hint");A([n()],$.prototype,"error");A([n()],$.prototype,"value");A([n({reflect:!0})],$.prototype,"name");A([n()],$.prototype,"placeholder");A([n({reflect:!0})],$.prototype,"type");A([n({type:Boolean,reflect:!0})],$.prototype,"disabled");A([n({type:Boolean,reflect:!0})],$.prototype,"required");A([n({type:Boolean,reflect:!0})],$.prototype,"invalid");A([n({reflect:!0})],$.prototype,"density");A([n({type:Boolean,reflect:!0,attribute:"hide-label"})],$.prototype,"hideLabel");A([n()],$.prototype,"min");A([n()],$.prototype,"max");A([n()],$.prototype,"step");A([n()],$.prototype,"accept");A([n({type:Boolean})],$.prototype,"multiple");b("mb-input",$);var Te=Object.defineProperty,C=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Te(t,e,s),s},w=class extends m{constructor(){super(...arguments),this.label="",this.hint="",this.error="",this.value="",this.name="",this.placeholder="",this.disabled=!1,this.required=!1,this.invalid=!1,this.rows=4,this.density="default",this.hideLabel=!1,this.#t=this.attachInternals(),this.#e=!1,this.#r="",this.#o=!1,this.#i=!1}static{this.formAssociated=!0}static{this.styles=[f,tt,u`
      :host {
        display: block;
      }

      textarea.control {
        min-block-size: 6rem;
        resize: vertical;
      }
    `]}#t;#e;#s;#r;#o;#i;get#d(){return this.disabled||this.#e}get#a(){return this.getAttribute("aria-label")??""}connectedCallback(){super.connectedCallback(),this.#o||(this.#r=this.value,this.#o=!0)}firstUpdated(){this.#s=this.renderRoot.querySelector("textarea")??void 0,this.#n()}updated(t){(t.has("value")||t.has("required")||t.has("error")||t.has("disabled")||t.has("name"))&&this.#n()}formDisabledCallback(t){this.#e=t,this.requestUpdate()}formResetCallback(){this.#i=!1,this.value=this.#r,this.error="",this.invalid=!1}#n(){k(this.#t,this.name?this.value:null);let t=this.required&&!this.value,{flags:e,message:i}=j(this.error,t);i?(M(this.#t,e,i,this.#s),this.invalid=!!this.error||this.#i):(q(this.#t),this.invalid=!1)}#l(t){let e=t.target;this.#i=!0,this.value=e.value,this.dispatchEvent(new CustomEvent("mb-input",{detail:{value:this.value},bubbles:!0,composed:!0}))}#h(t){let e=t.target;this.#i=!0,this.value=e.value,this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value},bubbles:!0,composed:!0}))}render(){let t=[this.hint&&!this.error?"hint":"",this.error?"error":""].filter(Boolean).join(" "),{labelText:e,hideVisually:i,controlAriaLabel:s}=et(this.label,this.hideLabel,this.#a);return d`
      <div class="field">
        ${e?d`<label
              part="label"
              class="label${i?" visually-hidden":""}"
              for="control"
              >${e}</label
            >`:l}
        <textarea
          id="control"
          part="control"
          class="control"
          .value=${this.value}
          name=${this.name||l}
          placeholder=${this.placeholder||l}
          rows=${this.rows}
          ?disabled=${this.#d}
          ?required=${this.required}
          aria-invalid=${this.invalid?"true":"false"}
          aria-label=${s||l}
          aria-describedby=${t||l}
          @input=${this.#l}
          @change=${this.#h}
        ></textarea>
        ${this.hint&&!this.error?d`<p id="hint" class="hint">${this.hint}</p>`:l}
        ${this.error?d`<p id="error" class="error" role="alert">${this.error}</p>`:l}
      </div>
    `}};C([n()],w.prototype,"label");C([n()],w.prototype,"hint");C([n()],w.prototype,"error");C([n()],w.prototype,"value");C([n({reflect:!0})],w.prototype,"name");C([n()],w.prototype,"placeholder");C([n({type:Boolean,reflect:!0})],w.prototype,"disabled");C([n({type:Boolean,reflect:!0})],w.prototype,"required");C([n({type:Boolean,reflect:!0})],w.prototype,"invalid");C([n({type:Number})],w.prototype,"rows");C([n({reflect:!0})],w.prototype,"density");C([n({type:Boolean,reflect:!0,attribute:"hide-label"})],w.prototype,"hideLabel");b("mb-textarea",w);var Ne=Object.defineProperty,B=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Ne(t,e,s),s},z=class extends m{constructor(){super(...arguments),this.label="",this.error="",this.name="",this.value="on",this.checked=!1,this.indeterminate=!1,this.disabled=!1,this.required=!1,this.invalid=!1,this.#t=this.attachInternals(),this.#e=!1,this.#r=!1,this.#o=!1,this.#i=!1}static{this.formAssociated=!0}static{this.styles=[f,u`
      :host {
        display: inline-block;
      }

      label {
        display: inline-flex;
        align-items: flex-start;
        gap: var(--mb-space-2);
        cursor: pointer;
        font-size: var(--mb-font-size-md);
      }

      input {
        margin-block-start: 0.2rem;
        accent-color: var(--mb-color-accent);
        inline-size: 1.1rem;
        block-size: 1.1rem;
      }

      input:disabled {
        cursor: not-allowed;
      }

      :host([disabled]) label {
        opacity: 0.55;
        cursor: not-allowed;
      }

      .error {
        margin: var(--mb-space-1) 0 0;
        color: var(--mb-color-danger);
        font-size: var(--mb-font-size-sm);
      }
    `]}#t;#e;#s;#r;#o;#i;get#d(){return this.disabled||this.#e}connectedCallback(){super.connectedCallback(),this.#o||(this.#r=this.checked,this.#o=!0)}firstUpdated(){this.#s=this.renderRoot.querySelector("input")??void 0,this.#a(),this.#n()}updated(t){t.has("indeterminate")&&this.#a(),(t.has("checked")||t.has("value")||t.has("required")||t.has("error")||t.has("disabled")||t.has("name"))&&this.#n()}formDisabledCallback(t){this.#e=t,this.requestUpdate()}formResetCallback(){this.#i=!1,this.checked=this.#r,this.indeterminate=!1,this.error="",this.invalid=!1}#a(){this.#s&&(this.#s.indeterminate=this.indeterminate)}#n(){k(this.#t,this.name&&this.checked?this.value:null);let t=this.required&&!this.checked,e=this.error||(t?"Please check this box.":"");if(e){let i=this.error?{customError:!0}:{valueMissing:!0};M(this.#t,i,e,this.#s),this.invalid=!!this.error||this.#i}else q(this.#t),this.invalid=!1}#l(t){let e=t.target;this.#i=!0,this.checked=e.checked,this.indeterminate=!1,this.dispatchEvent(new CustomEvent("mb-change",{detail:{checked:this.checked,value:this.value},bubbles:!0,composed:!0}))}render(){let t=this.error?"error":"";return d`
      <label part="label">
        <input
          part="control"
          type="checkbox"
          .checked=${this.checked}
          name=${this.name||l}
          value=${this.value}
          ?disabled=${this.#d}
          ?required=${this.required}
          aria-invalid=${this.invalid?"true":"false"}
          aria-describedby=${t||l}
          @change=${this.#l}
        />
        <span>${this.label}<slot></slot></span>
      </label>
      ${this.error?d`<p id="error" class="error" role="alert">${this.error}</p>`:l}
    `}};B([n()],z.prototype,"label");B([n()],z.prototype,"error");B([n({reflect:!0})],z.prototype,"name");B([n()],z.prototype,"value");B([n({type:Boolean,reflect:!0})],z.prototype,"checked");B([n({type:Boolean,reflect:!0})],z.prototype,"indeterminate");B([n({type:Boolean,reflect:!0})],z.prototype,"disabled");B([n({type:Boolean,reflect:!0})],z.prototype,"required");B([n({type:Boolean,reflect:!0})],z.prototype,"invalid");b("mb-checkbox",z);var he={ATTRIBUTE:1,CHILD:2,PROPERTY:3,BOOLEAN_ATTRIBUTE:4,EVENT:5,ELEMENT:6},de=r=>(...t)=>({_$litDirective$:r,values:t}),St=class{constructor(t){}get _$AU(){return this._$AM._$AU}_$AT(t,e,i){this._$Ct=t,this._$AM=e,this._$Ci=i}_$AS(t,e){return this.update(t,e)}update(t,e){return this.render(...e)}};var{I:Be}=ae,ce=r=>r;var pe=()=>document.createComment(""),st=(r,t,e)=>{let i=r._$AA.parentNode,s=t===void 0?r._$AB:t._$AA;if(e===void 0){let o=i.insertBefore(pe(),s),a=i.insertBefore(pe(),s);e=new Be(o,a,r,r.options)}else{let o=e._$AB.nextSibling,a=e._$AM,p=a!==r;if(p){let h;e._$AQ?.(r),e._$AM=r,e._$AP!==void 0&&(h=r._$AU)!==a._$AU&&e._$AP(h)}if(o!==s||p){let h=e._$AA;for(;h!==o;){let y=ce(h).nextSibling;ce(i).insertBefore(h,s),h=y}}}return e},H=(r,t,e=r)=>(r._$AI(t,e),r),je={},me=(r,t=je)=>r._$AH=t,ue=r=>r._$AH,Ct=r=>{r._$AR(),r._$AA.remove()};var be=(r,t,e)=>{let i=new Map;for(let s=t;s<=e;s++)i.set(r[s],s);return i},fe=de(class extends St{constructor(r){if(super(r),r.type!==he.CHILD)throw Error("repeat() can only be used in text expressions")}dt(r,t,e){let i;e===void 0?e=t:t!==void 0&&(i=t);let s=[],o=[],a=0;for(let p of r)s[a]=i?i(p,a):a,o[a]=e(p,a),a++;return{values:o,keys:s}}render(r,t,e){return this.dt(r,t,e).values}update(r,[t,e,i]){let s=ue(r),{values:o,keys:a}=this.dt(t,e,i);if(!Array.isArray(s))return this.ut=a,o;let p=this.ut??=[],h=[],y,x,c=0,g=s.length-1,v=0,_=o.length-1;for(;c<=g&&v<=_;)if(s[c]===null)c++;else if(s[g]===null)g--;else if(p[c]===a[v])h[v]=H(s[c],o[v]),c++,v++;else if(p[g]===a[_])h[_]=H(s[g],o[_]),g--,_--;else if(p[c]===a[_])h[_]=H(s[c],o[_]),st(r,h[_+1],s[c]),c++,_--;else if(p[g]===a[v])h[v]=H(s[g],o[v]),st(r,s[c],s[g]),g--,v++;else if(y===void 0&&(y=be(a,v,_),x=be(p,c,g)),y.has(p[c]))if(y.has(p[g])){let D=x.get(a[v]),Ut=D!==void 0?s[D]:null;if(Ut===null){let Vt=st(r,s[c]);H(Vt,o[v]),h[v]=Vt}else h[v]=H(Ut,o[v]),st(r,s[c],Ut),s[D]=null;v++}else Ct(s[g]),g--;else Ct(s[c]),c++;for(;v<=_;){let D=st(r,h[_+1]);H(D,o[v]),h[v++]=D}for(;c<=g;){let D=s[c++];D!==null&&Ct(D)}return this.ut=a,me(r,h),N}});var He=Object.defineProperty,P=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&He(t,e,s),s};function Ve(r){if(!r)return[];try{let t=JSON.parse(r);return Array.isArray(t)?t.filter(e=>!!e&&typeof e=="object"&&typeof e.value=="string"&&typeof e.label=="string").map(e=>({value:e.value,label:e.label,disabled:!!e.disabled})):[]}catch{return[]}}var E=class extends m{constructor(){super(...arguments),this.label="",this.hint="",this.error="",this.value="",this.name="",this.disabled=!1,this.required=!1,this.invalid=!1,this.density="default",this.hideLabel=!1,this.options=[],this._slottedOptions=[],this.#t=this.attachInternals(),this.#e=!1,this.#r="",this.#o=!1,this.#i=!1}static{this.formAssociated=!0}static{this.styles=[f,tt,u`
      :host {
        display: block;
      }

      slot[name='options'] {
        display: none;
      }
    `]}#t;#e;#s;#r;#o;#i;get#d(){return this.disabled||this.#e}get#a(){return this._slottedOptions.length?this._slottedOptions:this.options}get#n(){return this.getAttribute("aria-label")??""}connectedCallback(){super.connectedCallback(),this.#o||(this.#r=this.value,this.#o=!0),this.#h()}firstUpdated(){this.#s=this.renderRoot.querySelector("select")??void 0,this.#m()}updated(t){(t.has("value")||t.has("required")||t.has("error")||t.has("options")||t.has("_slottedOptions")||t.has("disabled")||t.has("name"))&&this.#m()}formDisabledCallback(t){this.#e=t,this.requestUpdate()}formResetCallback(){this.#i=!1,this.value=this.#r,this.error="",this.invalid=!1}#l(t){return t instanceof HTMLOptionElement?{value:t.value,label:t.label||t.textContent?.trim()||t.value,disabled:t.disabled}:null}#h(){let t=[...this.querySelectorAll(":scope > option")].map(e=>this.#l(e)).filter(e=>e!=null);t.length&&(this._slottedOptions=t)}#c(){let t=this.renderRoot.querySelector('slot[name="options"]'),e=this.renderRoot.querySelector("slot:not([name])"),i=[...t?.assignedElements({flatten:!0})??[],...e?.assignedElements({flatten:!0})??[]].map(a=>this.#l(a)).filter(a=>a!=null),s=JSON.stringify(this._slottedOptions),o=JSON.stringify(i);s!==o&&(this._slottedOptions=i)}#p(){this.#c()}#m(){this.#s&&this.#s.value!==this.value&&(this.#s.value=this.value),k(this.#t,this.name?this.value:null);let t=this.required&&!this.value,{flags:e,message:i}=j(this.error,t,"Please select an option.");i?(M(this.#t,e,i,this.#s),this.invalid=!!this.error||this.#i):(q(this.#t),this.invalid=!1)}#u(t){let e=t.target;this.#i=!0,this.value=e.value,this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value},bubbles:!0,composed:!0}))}render(){let t=[this.hint&&!this.error?"hint":"",this.error?"error":""].filter(Boolean).join(" "),{labelText:e,hideVisually:i,controlAriaLabel:s}=et(this.label,this.hideLabel,this.#n);return d`
      <div class="field">
        ${e?d`<label
              part="label"
              class="label${i?" visually-hidden":""}"
              for="control"
              >${e}</label
            >`:l}
        <select
          id="control"
          part="control"
          class="control"
          name=${this.name||l}
          ?disabled=${this.#d}
          ?required=${this.required}
          aria-invalid=${this.invalid?"true":"false"}
          aria-label=${s||l}
          aria-describedby=${t||l}
          .value=${this.value}
          @change=${this.#u}
        >
          <option value="" ?disabled=${this.required}></option>
          ${fe(this.#a,o=>o.value,o=>d`
              <option value=${o.value} ?disabled=${!!o.disabled}>
                ${o.label}
              </option>
            `)}
        </select>
        ${this.hint&&!this.error?d`<p id="hint" class="hint">${this.hint}</p>`:l}
        ${this.error?d`<p id="error" class="error" role="alert">${this.error}</p>`:l}
      </div>
      <slot name="options" @slotchange=${this.#p}></slot>
      <slot @slotchange=${this.#p}></slot>
    `}};P([n()],E.prototype,"label");P([n()],E.prototype,"hint");P([n()],E.prototype,"error");P([n()],E.prototype,"value");P([n({reflect:!0})],E.prototype,"name");P([n({type:Boolean,reflect:!0})],E.prototype,"disabled");P([n({type:Boolean,reflect:!0})],E.prototype,"required");P([n({type:Boolean,reflect:!0})],E.prototype,"invalid");P([n({reflect:!0})],E.prototype,"density");P([n({type:Boolean,reflect:!0,attribute:"hide-label"})],E.prototype,"hideLabel");P([n({attribute:"options",converter:{fromAttribute:Ve,toAttribute(r){return r?.length?JSON.stringify(r):null}}})],E.prototype,"options");P([dt()],E.prototype,"_slottedOptions");b("mb-select",E);var Ie=Object.defineProperty,ve=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Ie(t,e,s),s},pt=class extends m{constructor(){super(...arguments),this.open=!1,this.heading="",this.#e=!1}static{this.styles=[f,u`
      :host {
        display: contents;
      }

      dialog {
        border: 1px solid var(--mb-color-border);
        border-radius: var(--mb-radius-lg);
        padding: 0;
        background: var(--mb-color-surface);
        color: var(--mb-color-fg);
        box-shadow: var(--mb-shadow);
        /* Avoid 100vw — it includes scrollbar gutters and overflows on mobile */
        inline-size: min(32rem, calc(100% - 2rem));
        max-inline-size: calc(100% - 2rem);
        margin: auto;
      }

      dialog::backdrop {
        background: rgb(20 32 27 / 45%);
      }

      .panel {
        display: flex;
        flex-direction: column;
        gap: var(--mb-space-4);
        padding: var(--mb-space-5);
        min-inline-size: 0;
        max-inline-size: 100%;
      }

      .header {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: var(--mb-space-3);
        min-inline-size: 0;
      }

      .title {
        font-family: var(--mb-font-display);
        font-size: var(--mb-font-size-xl);
        font-weight: 650;
        margin: 0;
        min-inline-size: 0;
        flex: 1;
        overflow-wrap: anywhere;
      }

      .close {
        border: 0;
        background: transparent;
        color: var(--mb-color-muted);
        font-size: 1.25rem;
        line-height: 1;
        cursor: pointer;
        padding: var(--mb-space-1);
        border-radius: var(--mb-radius-sm);
        flex-shrink: 0;
      }
    `]}#t;#e;firstUpdated(){this.#t=this.renderRoot.querySelector("dialog")??void 0,this.#t?.addEventListener("close",()=>{this.#e||(this.open&&(this.open=!1),this.#r())}),this.#s()}updated(t){t.has("open")&&this.#s()}#s(){let t=this.#t;t&&(this.open&&!t.open?t.showModal():!this.open&&t.open&&(this.#e=!0,t.close(),this.#e=!1,this.#r()))}#r(){this.dispatchEvent(new CustomEvent("mb-close",{bubbles:!0,composed:!0}))}close(){!this.open&&!this.#t?.open||(this.open=!1)}#o(){this.close()}render(){return d`
      <dialog part="dialog" aria-labelledby="title" aria-modal="true">
        <div class="panel">
          <div class="header">
            <h2 class="title" id="title">${this.heading}<slot name="heading"></slot></h2>
            <button class="close" type="button" aria-label="Close" @click=${this.#o}>
              ×
            </button>
          </div>
          <div part="body">
            <slot></slot>
          </div>
          <div part="footer">
            <slot name="footer"></slot>
          </div>
        </div>
      </dialog>
    `}};ve([n({type:Boolean,reflect:!0})],pt.prototype,"open");ve([n()],pt.prototype,"heading");b("mb-modal",pt);var Fe=Object.defineProperty,zt=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Fe(t,e,s),s},Y=class extends m{constructor(){super(...arguments),this.value=0,this.max=100,this.percent=null,this.label=""}static{this.styles=[f,u`
      :host {
        display: block;
        inline-size: 100%;
      }

      .wrap {
        display: flex;
        flex-direction: column;
        gap: var(--mb-space-1);
      }

      .label {
        font-size: var(--mb-font-size-sm);
        color: var(--mb-color-muted);
      }

      .track {
        inline-size: 100%;
        block-size: 0.5rem;
        border-radius: var(--mb-radius-sm);
        background: var(--mb-color-border);
        overflow: clip;
      }

      .bar {
        block-size: 100%;
        background: var(--mb-color-accent);
        border-radius: inherit;
        transition: inline-size var(--mb-transition);
      }
    `]}get#t(){if(this.percent!=null&&!Number.isNaN(this.percent))return Math.min(100,Math.max(0,this.percent));let t=this.max>0?this.max:100;return Math.min(100,Math.max(0,this.value/t*100))}get#e(){return this.percent!=null&&!Number.isNaN(this.percent)?this.#t:this.value}get#s(){return this.percent!=null&&!Number.isNaN(this.percent)?100:this.max>0?this.max:100}render(){let t=this.#t;return d`
      <div class="wrap">
        ${this.label?d`<div part="label" class="label" id="label">${this.label}</div>`:l}
        <div
          part="track"
          class="track"
          role="progressbar"
          aria-valuemin="0"
          aria-valuenow=${this.#e}
          aria-valuemax=${this.#s}
          aria-labelledby=${this.label?"label":l}
          aria-label=${this.label?l:this.getAttribute("aria-label")||"Progress"}
        >
          <div part="bar" class="bar" style="inline-size: ${t}%"></div>
        </div>
        <slot></slot>
      </div>
    `}};zt([n({type:Number})],Y.prototype,"value");zt([n({type:Number})],Y.prototype,"max");zt([n({type:Number})],Y.prototype,"percent");zt([n()],Y.prototype,"label");b("mb-progress",Y);var Je=Object.defineProperty,We=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Je(t,e,s),s},Pt=class extends m{constructor(){super(...arguments),this.label="Filters"}static{this.styles=[f,u`
      :host {
        display: block;
        max-inline-size: 100%;
      }

      .scroller {
        overflow-x: auto;
        -webkit-overflow-scrolling: touch;
        max-inline-size: 100%;
      }

      .list {
        display: inline-flex;
        min-inline-size: 100%;
        gap: 0;
        padding: var(--mb-space-1);
        border: 1px solid var(--mb-color-border);
        border-radius: var(--mb-radius-md);
        background: var(--mb-color-surface);
      }

      ::slotted(a),
      ::slotted(button) {
        appearance: none;
        border: 0;
        background: transparent;
        color: var(--mb-color-muted);
        font: inherit;
        font-weight: 600;
        font-size: var(--mb-font-size-sm);
        text-decoration: none;
        padding-block: var(--mb-space-2);
        padding-inline: var(--mb-space-3);
        border-radius: var(--mb-radius-sm);
        white-space: nowrap;
        cursor: pointer;
      }

      ::slotted(a:focus-visible),
      ::slotted(button:focus-visible) {
        outline: var(--mb-focus-ring);
        outline-offset: var(--mb-focus-offset);
      }

      ::slotted([aria-current='page']),
      ::slotted([aria-selected='true']),
      ::slotted(.is-active) {
        background: var(--mb-color-accent-soft);
        color: var(--mb-color-accent);
      }
    `]}render(){return d`
      <nav part="nav" class="scroller" aria-label=${this.label}>
        <div part="list" class="list" role="list">
          <slot></slot>
        </div>
      </nav>
    `}};We([n()],Pt.prototype,"label");b("mb-segmented-control",Pt);var Ke=Object.defineProperty,Ge=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Ke(t,e,s),s},Ot=class extends m{constructor(){super(...arguments),this.heading=""}static{this.styles=[f,u`
      :host {
        display: block;
        inline-size: 100%;
      }

      .panel {
        display: flex;
        flex-direction: column;
        align-items: flex-start;
        gap: var(--mb-space-3);
        padding-block: var(--mb-space-6);
        padding-inline: var(--mb-space-5);
        border: 1px dashed var(--mb-color-border);
        border-radius: var(--mb-radius-lg);
        background: var(--mb-color-surface);
      }

      .heading {
        margin: 0;
        font-family: var(--mb-font-display);
        font-size: var(--mb-font-size-lg);
        font-weight: 650;
        line-height: var(--mb-line-height-tight);
        color: var(--mb-color-fg);
      }

      .body {
        color: var(--mb-color-muted);
        font-size: var(--mb-font-size-md);
      }

      .actions {
        display: flex;
        flex-wrap: wrap;
        gap: var(--mb-space-2);
      }

      .actions:not([data-has-content]) {
        display: none;
      }
    `]}#t(t){let e=t.target.assignedNodes({flatten:!0}).length>0;this.renderRoot.querySelector(".actions")?.toggleAttribute("data-has-content",e)}render(){return d`
      <div part="panel" class="panel">
        ${this.heading?d`<h2 part="heading" class="heading">${this.heading}</h2>`:d`<slot name="heading"></slot>`}
        <div part="body" class="body">
          <slot></slot>
        </div>
        <div part="actions" class="actions">
          <slot name="actions" @slotchange=${this.#t}></slot>
        </div>
      </div>
    `}};Ge([n()],Ot.prototype,"heading");b("mb-empty-state",Ot);var Qe=Object.defineProperty,V=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Qe(t,e,s),s},L=class extends m{constructor(){super(...arguments),this.prevUrl="",this.nextUrl="",this.prevDisabled=!1,this.nextDisabled=!1,this.status="",this.prevLabel="Previous",this.nextLabel="Next",this.label="Pagination"}static{this.styles=[f,u`
      :host {
        display: block;
      }

      nav {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        justify-content: space-between;
        gap: var(--mb-space-3);
      }

      .status {
        color: var(--mb-color-muted);
        font-size: var(--mb-font-size-sm);
      }

      .actions {
        display: inline-flex;
        gap: var(--mb-space-2);
      }

      a,
      span.disabled {
        display: inline-flex;
        align-items: center;
        min-block-size: 2.25rem;
        padding-inline: var(--mb-space-3);
        border: 1px solid var(--mb-color-border);
        border-radius: var(--mb-radius-md);
        background: var(--mb-color-surface);
        color: var(--mb-color-fg);
        font-size: var(--mb-font-size-sm);
        font-weight: 600;
        text-decoration: none;
      }

      span.disabled {
        opacity: 0.45;
        cursor: not-allowed;
      }
    `]}render(){let t=this.prevDisabled||!this.prevUrl,e=this.nextDisabled||!this.nextUrl;return d`
      <nav part="nav" aria-label=${this.label}>
        <div part="status" class="status">${this.status}<slot name="status"></slot></div>
        <div part="actions" class="actions">
          <slot name="prev">
            ${t?d`<span class="disabled" aria-disabled="true">${this.prevLabel}</span>`:d`<a part="prev" href=${this.prevUrl}>${this.prevLabel}</a>`}
          </slot>
          <slot name="next">
            ${e?d`<span class="disabled" aria-disabled="true">${this.nextLabel}</span>`:d`<a part="next" href=${this.nextUrl}>${this.nextLabel}</a>`}
          </slot>
        </div>
      </nav>
    `}};V([n({attribute:"prev-url"})],L.prototype,"prevUrl");V([n({attribute:"next-url"})],L.prototype,"nextUrl");V([n({type:Boolean,attribute:"prev-disabled"})],L.prototype,"prevDisabled");V([n({type:Boolean,attribute:"next-disabled"})],L.prototype,"nextDisabled");V([n()],L.prototype,"status");V([n({attribute:"prev-label"})],L.prototype,"prevLabel");V([n({attribute:"next-label"})],L.prototype,"nextLabel");V([n()],L.prototype,"label");b("mb-pagination",L);var Ye=Object.defineProperty,Lt=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Ye(t,e,s),s},Z=class extends m{constructor(){super(...arguments),this.open=!1,this.variant="info",this.autoDismiss=4e3,this.message="",this.#t=0,this.#e=t=>{let e=t.detail;e&&(e.variant&&(this.variant=e.variant),e.message!=null&&(this.message=e.message),e.autoDismiss!=null&&(this.autoDismiss=e.autoDismiss),this.show())}}static{this.styles=[f,u`
      :host {
        display: block;
        position: fixed;
        inset-block-end: var(--mb-space-5);
        inset-inline: var(--mb-space-4);
        z-index: 1000;
        pointer-events: none;
      }

      :host(:not([open])) {
        visibility: hidden;
      }

      .toast {
        pointer-events: auto;
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: var(--mb-space-3);
        max-inline-size: 28rem;
        margin-inline: auto;
        padding-block: var(--mb-space-3);
        padding-inline: var(--mb-space-4);
        border-radius: var(--mb-radius-md);
        border: 1px solid var(--mb-color-border);
        background: var(--mb-color-surface);
        box-shadow: var(--mb-shadow);
        color: var(--mb-color-fg);
      }

      :host([variant='success']) .toast {
        border-color: var(--mb-color-success);
        background: var(--mb-color-success-soft);
        color: var(--mb-color-success);
      }

      :host([variant='danger']) .toast {
        border-color: var(--mb-color-danger);
        background: var(--mb-color-danger-soft);
        color: var(--mb-color-danger);
      }

      :host([variant='info']) .toast {
        border-color: var(--mb-color-info);
        background: var(--mb-color-info-soft);
        color: var(--mb-color-info);
      }

      .message {
        flex: 1;
        font-size: var(--mb-font-size-sm);
        font-weight: 600;
      }

      button {
        appearance: none;
        border: 0;
        background: transparent;
        color: inherit;
        cursor: pointer;
        font: inherit;
        font-weight: 700;
        line-height: 1;
        padding: 0;
      }
    `]}#t;#e;connectedCallback(){super.connectedCallback(),document.addEventListener("mb-toast",this.#e)}disconnectedCallback(){super.disconnectedCallback(),document.removeEventListener("mb-toast",this.#e),this.#r()}updated(t){t.has("open")&&(this.open?this.#s():this.#r())}show(t,e){t!=null&&(this.message=t),e&&(this.variant=e),this.open=!0}hide(){this.open=!1}#s(){this.#r(),this.autoDismiss>0&&(this.#t=window.setTimeout(()=>this.hide(),this.autoDismiss))}#r(){this.#t&&(window.clearTimeout(this.#t),this.#t=0)}#o(){this.hide(),this.dispatchEvent(new CustomEvent("mb-close",{bubbles:!0,composed:!0}))}render(){let t=this.variant==="danger"?"alert":"status";return d`
      <div
        part="toast"
        class="toast"
        role=${t}
        aria-live=${this.variant==="danger"?"assertive":"polite"}
        ?hidden=${!this.open}
      >
        <div part="message" class="message">${this.message}<slot></slot></div>
        <button type="button" part="close" aria-label="Dismiss" @click=${this.#o}>
          ×
        </button>
      </div>
    `}};Lt([n({type:Boolean,reflect:!0})],Z.prototype,"open");Lt([n({reflect:!0})],Z.prototype,"variant");Lt([n({type:Number,attribute:"auto-dismiss"})],Z.prototype,"autoDismiss");Lt([n()],Z.prototype,"message");b("mb-toast",Z);var Ze=Object.defineProperty,mt=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Ze(t,e,s),s},I=class extends m{constructor(){super(...arguments),this.value="",this.label="",this.disabled=!1,this.checked=!1,this.name=""}static{this.styles=[f,u`
      :host {
        display: block;
      }

      label {
        display: inline-flex;
        align-items: flex-start;
        gap: var(--mb-space-2);
        cursor: pointer;
        font-size: var(--mb-font-size-md);
      }

      input {
        margin-block-start: 0.2rem;
        accent-color: var(--mb-color-accent);
      }

      :host([disabled]) label {
        opacity: 0.55;
        cursor: not-allowed;
      }
    `]}#t;firstUpdated(){this.#t=this.renderRoot.querySelector("input")??void 0}focus(t){this.#t?.focus(t)}#e(){this.checked=!0,this.dispatchEvent(new CustomEvent("mb-radio-select",{detail:{value:this.value},bubbles:!0,composed:!0}))}render(){return d`
      <label part="label">
        <input
          part="control"
          type="radio"
          name=${this.name||l}
          .value=${this.value}
          .checked=${this.checked}
          ?disabled=${this.disabled}
          @change=${this.#e}
        />
        <span>${this.label}<slot></slot></span>
      </label>
    `}};mt([n()],I.prototype,"value");mt([n()],I.prototype,"label");mt([n({type:Boolean,reflect:!0})],I.prototype,"disabled");mt([n({type:Boolean,reflect:!0})],I.prototype,"checked");mt([n()],I.prototype,"name");b("mb-radio",I);var Xe=Object.defineProperty,F=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Xe(t,e,s),s};function ts(r){if(!r)return[];try{let t=JSON.parse(r);return Array.isArray(t)?t.filter(e=>!!e&&typeof e=="object"&&typeof e.value=="string"&&typeof e.label=="string").map(e=>({value:e.value,label:e.label,disabled:!!e.disabled})):[]}catch{return[]}}var U=class extends m{constructor(){super(...arguments),this.label="",this.error="",this.value="",this.name="",this.disabled=!1,this.required=!1,this.invalid=!1,this.options=[],this.#t=this.attachInternals(),this.#e=!1,this.#s="",this.#r=!1,this.#o=!1,this.#l=t=>{let e=t.detail?.value;e!=null&&(this.#o=!0,this.value=e,this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value},bubbles:!0,composed:!0})))},this.#h=t=>{if(!["ArrowDown","ArrowUp","ArrowRight","ArrowLeft"].includes(t.key))return;let e=this.#d().filter(a=>!a.disabled);if(!e.length)return;t.preventDefault();let i=e.findIndex(a=>a.value===this.value),s=t.key==="ArrowDown"||t.key==="ArrowRight"?1:-1,o=e[(i+s+e.length)%e.length];this.#o=!0,this.value=o.value,o.focus(),this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value},bubbles:!0,composed:!0}))}}static{this.formAssociated=!0}static{this.styles=[f,u`
      :host {
        display: block;
      }

      fieldset {
        margin: 0;
        padding: 0;
        border: 0;
        min-inline-size: 0;
      }

      legend {
        font-size: var(--mb-font-size-sm);
        font-weight: 600;
        margin-block-end: var(--mb-space-2);
      }

      .options {
        display: flex;
        flex-direction: column;
        gap: var(--mb-space-2);
      }

      .error {
        margin: var(--mb-space-2) 0 0;
        color: var(--mb-color-danger);
        font-size: var(--mb-font-size-sm);
      }
    `]}#t;#e;#s;#r;#o;get#i(){return this.disabled||this.#e}connectedCallback(){super.connectedCallback(),this.#r||(this.#s=this.value,this.#r=!0),this.addEventListener("mb-radio-select",this.#l),this.addEventListener("keydown",this.#h)}disconnectedCallback(){super.disconnectedCallback(),this.removeEventListener("mb-radio-select",this.#l),this.removeEventListener("keydown",this.#h)}firstUpdated(){this.#a(),this.#n()}updated(t){(t.has("value")||t.has("name")||t.has("disabled")||t.has("options"))&&this.#a(),(t.has("value")||t.has("required")||t.has("error")||t.has("name")||t.has("disabled"))&&this.#n()}formDisabledCallback(t){this.#e=t,this.requestUpdate(),this.#a()}formResetCallback(){this.#o=!1,this.value=this.#s,this.error="",this.invalid=!1}#d(){let t=this.renderRoot.querySelector("slot")?.assignedElements({flatten:!0}).filter(i=>i.localName==="mb-radio")??[],e=[...this.renderRoot.querySelectorAll(".options > mb-radio")];return[...t,...e]}#a(){let t=this.#d();for(let e of t)e.name=this.name||"mb-radio-group",e.checked=e.value===this.value,this.#i&&(e.disabled=!0)}#n(){k(this.#t,this.name?this.value:null);let t=this.required&&!this.value,{flags:e,message:i}=j(this.error,t,"Please select an option.");i?(M(this.#t,e,i),this.invalid=!!this.error||this.#o):(q(this.#t),this.invalid=!1)}#l;#h;#c(){this.#a()}render(){return d`
      <fieldset part="fieldset" ?disabled=${this.#i}>
        ${this.label?d`<legend part="legend">${this.label}</legend>`:l}
        <div class="options" part="options" role="radiogroup" aria-invalid=${this.invalid?"true":"false"}>
          <slot @slotchange=${this.#c}></slot>
          ${this.options.map(t=>d`
              <mb-radio
                .value=${t.value}
                .label=${t.label}
                ?disabled=${!!t.disabled||this.#i}
                ?checked=${t.value===this.value}
                .name=${this.name||"mb-radio-group"}
              ></mb-radio>
            `)}
        </div>
        ${this.error?d`<p class="error" role="alert">${this.error}</p>`:l}
      </fieldset>
    `}};F([n()],U.prototype,"label");F([n()],U.prototype,"error");F([n()],U.prototype,"value");F([n({reflect:!0})],U.prototype,"name");F([n({type:Boolean,reflect:!0})],U.prototype,"disabled");F([n({type:Boolean,reflect:!0})],U.prototype,"required");F([n({type:Boolean,reflect:!0})],U.prototype,"invalid");F([n({attribute:"options",converter:{fromAttribute:ts,toAttribute(r){return r?.length?JSON.stringify(r):null}}})],U.prototype,"options");b("mb-radio-group",U);
