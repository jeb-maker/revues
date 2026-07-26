var xt=globalThis,_t=xt.ShadowRoot&&(xt.ShadyCSS===void 0||xt.ShadyCSS.nativeShadow)&&"adoptedStyleSheets"in Document.prototype&&"replace"in CSSStyleSheet.prototype,Vt=Symbol(),te=new WeakMap,at=class{constructor(t,e,i){if(this._$cssResult$=!0,i!==Vt)throw Error("CSSResult is not constructable. Use `unsafeCSS` or `css` instead.");this.cssText=t,this.t=e}get styleSheet(){let t=this.o,e=this.t;if(_t&&t===void 0){let i=e!==void 0&&e.length===1;i&&(t=te.get(e)),t===void 0&&((this.o=t=new CSSStyleSheet).replaceSync(this.cssText),i&&te.set(e,t))}return t}toString(){return this.cssText}},ee=r=>new at(typeof r=="string"?r:r+"",void 0,Vt),p=(r,...t)=>{let e=r.length===1?r[0]:t.reduce((i,s,o)=>i+(a=>{if(a._$cssResult$===!0)return a.cssText;if(typeof a=="number")return a;throw Error("Value passed to 'css' function must be a 'css' function result: "+a+". Use 'unsafeCSS' to pass non-literal values, but take care to ensure page security.")})(s)+r[o+1],r[0]);return new at(e,r,Vt)},se=(r,t)=>{if(_t)r.adoptedStyleSheets=t.map(e=>e instanceof CSSStyleSheet?e:e.styleSheet);else for(let e of t){let i=document.createElement("style"),s=xt.litNonce;s!==void 0&&i.setAttribute("nonce",s),i.textContent=e.cssText,r.appendChild(i)}},It=_t?r=>r:r=>r instanceof CSSStyleSheet?(t=>{let e="";for(let i of t.cssRules)e+=i.cssText;return ee(e)})(r):r;var{is:Oe,defineProperty:Le,getOwnPropertyDescriptor:Me,getOwnPropertyNames:De,getOwnPropertySymbols:Ue,getPrototypeOf:qe}=Object,At=globalThis,re=At.trustedTypes,Te=re?re.emptyScript:"",Ne=At.reactiveElementPolyfillSupport,nt=(r,t)=>r,lt={toAttribute(r,t){switch(t){case Boolean:r=r?Te:null;break;case Object:case Array:r=r==null?r:JSON.stringify(r)}return r},fromAttribute(r,t){let e=r;switch(t){case Boolean:e=r!==null;break;case Number:e=r===null?null:Number(r);break;case Object:case Array:try{e=JSON.parse(r)}catch{e=null}}return e}},wt=(r,t)=>!Oe(r,t),ie={attribute:!0,type:String,converter:lt,reflect:!1,useDefault:!1,hasChanged:wt};Symbol.metadata??=Symbol("metadata"),At.litPropertyMetadata??=new WeakMap;var T=class extends HTMLElement{static addInitializer(t){this._$Ei(),(this.l??=[]).push(t)}static get observedAttributes(){return this.finalize(),this._$Eh&&[...this._$Eh.keys()]}static createProperty(t,e=ie){if(e.state&&(e.attribute=!1),this._$Ei(),this.prototype.hasOwnProperty(t)&&((e=Object.create(e)).wrapped=!0),this.elementProperties.set(t,e),!e.noAccessor){let i=Symbol(),s=this.getPropertyDescriptor(t,i,e);s!==void 0&&Le(this.prototype,t,s)}}static getPropertyDescriptor(t,e,i){let{get:s,set:o}=Me(this.prototype,t)??{get(){return this[e]},set(a){this[e]=a}};return{get:s,set(a){let f=s?.call(this);o?.call(this,a),this.requestUpdate(t,f,i)},configurable:!0,enumerable:!0}}static getPropertyOptions(t){return this.elementProperties.get(t)??ie}static _$Ei(){if(this.hasOwnProperty(nt("elementProperties")))return;let t=qe(this);t.finalize(),t.l!==void 0&&(this.l=[...t.l]),this.elementProperties=new Map(t.elementProperties)}static finalize(){if(this.hasOwnProperty(nt("finalized")))return;if(this.finalized=!0,this._$Ei(),this.hasOwnProperty(nt("properties"))){let e=this.properties,i=[...De(e),...Ue(e)];for(let s of i)this.createProperty(s,e[s])}let t=this[Symbol.metadata];if(t!==null){let e=litPropertyMetadata.get(t);if(e!==void 0)for(let[i,s]of e)this.elementProperties.set(i,s)}this._$Eh=new Map;for(let[e,i]of this.elementProperties){let s=this._$Eu(e,i);s!==void 0&&this._$Eh.set(s,e)}this.elementStyles=this.finalizeStyles(this.styles)}static finalizeStyles(t){let e=[];if(Array.isArray(t)){let i=new Set(t.flat(1/0).reverse());for(let s of i)e.unshift(It(s))}else t!==void 0&&e.push(It(t));return e}static _$Eu(t,e){let i=e.attribute;return i===!1?void 0:typeof i=="string"?i:typeof t=="string"?t.toLowerCase():void 0}constructor(){super(),this._$Ep=void 0,this.isUpdatePending=!1,this.hasUpdated=!1,this._$Em=null,this._$Ev()}_$Ev(){this._$ES=new Promise(t=>this.enableUpdating=t),this._$AL=new Map,this._$E_(),this.requestUpdate(),this.constructor.l?.forEach(t=>t(this))}addController(t){(this._$EO??=new Set).add(t),this.renderRoot!==void 0&&this.isConnected&&t.hostConnected?.()}removeController(t){this._$EO?.delete(t)}_$E_(){let t=new Map,e=this.constructor.elementProperties;for(let i of e.keys())this.hasOwnProperty(i)&&(t.set(i,this[i]),delete this[i]);t.size>0&&(this._$Ep=t)}createRenderRoot(){let t=this.shadowRoot??this.attachShadow(this.constructor.shadowRootOptions);return se(t,this.constructor.elementStyles),t}connectedCallback(){this.renderRoot??=this.createRenderRoot(),this.enableUpdating(!0),this._$EO?.forEach(t=>t.hostConnected?.())}enableUpdating(t){}disconnectedCallback(){this._$EO?.forEach(t=>t.hostDisconnected?.())}attributeChangedCallback(t,e,i){this._$AK(t,i)}_$ET(t,e){let i=this.constructor.elementProperties.get(t),s=this.constructor._$Eu(t,i);if(s!==void 0&&i.reflect===!0){let o=(i.converter?.toAttribute!==void 0?i.converter:lt).toAttribute(e,i.type);this._$Em=t,o==null?this.removeAttribute(s):this.setAttribute(s,o),this._$Em=null}}_$AK(t,e){let i=this.constructor,s=i._$Eh.get(t);if(s!==void 0&&this._$Em!==s){let o=i.getPropertyOptions(s),a=typeof o.converter=="function"?{fromAttribute:o.converter}:o.converter?.fromAttribute!==void 0?o.converter:lt;this._$Em=s;let f=a.fromAttribute(e,o.type);this[s]=f??this._$Ej?.get(s)??f,this._$Em=null}}requestUpdate(t,e,i,s=!1,o){if(t!==void 0){let a=this.constructor;if(s===!1&&(o=this[t]),i??=a.getPropertyOptions(t),!((i.hasChanged??wt)(o,e)||i.useDefault&&i.reflect&&o===this._$Ej?.get(t)&&!this.hasAttribute(a._$Eu(t,i))))return;this.C(t,e,i)}this.isUpdatePending===!1&&(this._$ES=this._$EP())}C(t,e,{useDefault:i,reflect:s,wrapped:o},a){i&&!(this._$Ej??=new Map).has(t)&&(this._$Ej.set(t,a??e??this[t]),o!==!0||a!==void 0)||(this._$AL.has(t)||(this.hasUpdated||i||(e=void 0),this._$AL.set(t,e)),s===!0&&this._$Em!==t&&(this._$Eq??=new Set).add(t))}async _$EP(){this.isUpdatePending=!0;try{await this._$ES}catch(e){Promise.reject(e)}let t=this.scheduleUpdate();return t!=null&&await t,!this.isUpdatePending}scheduleUpdate(){return this.performUpdate()}performUpdate(){if(!this.isUpdatePending)return;if(!this.hasUpdated){if(this.renderRoot??=this.createRenderRoot(),this._$Ep){for(let[s,o]of this._$Ep)this[s]=o;this._$Ep=void 0}let i=this.constructor.elementProperties;if(i.size>0)for(let[s,o]of i){let{wrapped:a}=o,f=this[s];a!==!0||this._$AL.has(s)||f===void 0||this.C(s,void 0,o,f)}}let t=!1,e=this._$AL;try{t=this.shouldUpdate(e),t?(this.willUpdate(e),this._$EO?.forEach(i=>i.hostUpdate?.()),this.update(e)):this._$EM()}catch(i){throw t=!1,this._$EM(),i}t&&this._$AE(e)}willUpdate(t){}_$AE(t){this._$EO?.forEach(e=>e.hostUpdated?.()),this.hasUpdated||(this.hasUpdated=!0,this.firstUpdated(t)),this.updated(t)}_$EM(){this._$AL=new Map,this.isUpdatePending=!1}get updateComplete(){return this.getUpdateComplete()}getUpdateComplete(){return this._$ES}shouldUpdate(t){return!0}update(t){this._$Eq&&=this._$Eq.forEach(e=>this._$ET(e,this[e])),this._$EM()}updated(t){}firstUpdated(t){}};T.elementStyles=[],T.shadowRootOptions={mode:"open"},T[nt("elementProperties")]=new Map,T[nt("finalized")]=new Map,Ne?.({ReactiveElement:T}),(At.reactiveElementVersions??=[]).push("2.1.2");var Jt=globalThis,oe=r=>r,kt=Jt.trustedTypes,ae=kt?kt.createPolicy("lit-html",{createHTML:r=>r}):void 0,Wt="$lit$",N=`lit$${Math.random().toFixed(9).slice(2)}$`,Kt="?"+N,Re=`<${Kt}>`,G=document,ct=()=>G.createComment(""),dt=r=>r===null||typeof r!="object"&&typeof r!="function",Gt=Array.isArray,pe=r=>Gt(r)||typeof r?.[Symbol.iterator]=="function",Ft=`[ 	
\f\r]`,ht=/<(?:(!--|\/[^a-zA-Z])|(\/?[a-zA-Z][^>\s]*)|(\/?$))/g,ne=/-->/g,le=/>/g,W=RegExp(`>|${Ft}(?:([^\\s"'>=/]+)(${Ft}*=${Ft}*(?:[^ 	
\f\r"'\`<>=]|("|')|))|$)`,"g"),he=/'/g,ce=/"/g,me=/^(?:script|style|textarea|title)$/i,Qt=r=>(t,...e)=>({_$litType$:r,strings:t,values:e}),l=Qt(1),ks=Qt(2),Es=Qt(3),R=Symbol.for("lit-noChange"),h=Symbol.for("lit-nothing"),de=new WeakMap,K=G.createTreeWalker(G,129);function ue(r,t){if(!Gt(r)||!r.hasOwnProperty("raw"))throw Error("invalid template strings array");return ae!==void 0?ae.createHTML(t):t}var be=(r,t)=>{let e=r.length-1,i=[],s,o=t===2?"<svg>":t===3?"<math>":"",a=ht;for(let f=0;f<e;f++){let c=r[f],y,x,b=-1,g=0;for(;g<c.length&&(a.lastIndex=g,x=a.exec(c),x!==null);)g=a.lastIndex,a===ht?x[1]==="!--"?a=ne:x[1]!==void 0?a=le:x[2]!==void 0?(me.test(x[2])&&(s=RegExp("</"+x[2],"g")),a=W):x[3]!==void 0&&(a=W):a===W?x[0]===">"?(a=s??ht,b=-1):x[1]===void 0?b=-2:(b=a.lastIndex-x[2].length,y=x[1],a=x[3]===void 0?W:x[3]==='"'?ce:he):a===ce||a===he?a=W:a===ne||a===le?a=ht:(a=W,s=void 0);let v=a===W&&r[f+1].startsWith("/>")?" ":"";o+=a===ht?c+Re:b>=0?(i.push(y),c.slice(0,b)+Wt+c.slice(b)+N+v):c+N+(b===-2?f:v)}return[ue(r,o+(r[e]||"<?>")+(t===2?"</svg>":t===3?"</math>":"")),i]},pt=class r{constructor({strings:t,_$litType$:e},i){let s;this.parts=[];let o=0,a=0,f=t.length-1,c=this.parts,[y,x]=be(t,e);if(this.el=r.createElement(y,i),K.currentNode=this.el.content,e===2||e===3){let b=this.el.content.firstChild;b.replaceWith(...b.childNodes)}for(;(s=K.nextNode())!==null&&c.length<f;){if(s.nodeType===1){if(s.hasAttributes())for(let b of s.getAttributeNames())if(b.endsWith(Wt)){let g=x[a++],v=s.getAttribute(b).split(N),_=/([.?@])?(.*)/.exec(g);c.push({type:1,index:o,name:_[2],strings:v,ctor:_[1]==="."?St:_[1]==="?"?zt:_[1]==="@"?Ct:Y}),s.removeAttribute(b)}else b.startsWith(N)&&(c.push({type:6,index:o}),s.removeAttribute(b));if(me.test(s.tagName)){let b=s.textContent.split(N),g=b.length-1;if(g>0){s.textContent=kt?kt.emptyScript:"";for(let v=0;v<g;v++)s.append(b[v],ct()),K.nextNode(),c.push({type:2,index:++o});s.append(b[g],ct())}}}else if(s.nodeType===8)if(s.data===Kt)c.push({type:2,index:o});else{let b=-1;for(;(b=s.data.indexOf(N,b+1))!==-1;)c.push({type:7,index:o}),b+=N.length-1}o++}}static createElement(t,e){let i=G.createElement("template");return i.innerHTML=t,i}};function Q(r,t,e=r,i){if(t===R)return t;let s=i!==void 0?e._$Co?.[i]:e._$Cl,o=dt(t)?void 0:t._$litDirective$;return s?.constructor!==o&&(s?._$AO?.(!1),o===void 0?s=void 0:(s=new o(r),s._$AT(r,e,i)),i!==void 0?(e._$Co??=[])[i]=s:e._$Cl=s),s!==void 0&&(t=Q(r,s._$AS(r,t.values),s,i)),t}var Et=class{constructor(t,e){this._$AV=[],this._$AN=void 0,this._$AD=t,this._$AM=e}get parentNode(){return this._$AM.parentNode}get _$AU(){return this._$AM._$AU}u(t){let{el:{content:e},parts:i}=this._$AD,s=(t?.creationScope??G).importNode(e,!0);K.currentNode=s;let o=K.nextNode(),a=0,f=0,c=i[0];for(;c!==void 0;){if(a===c.index){let y;c.type===2?y=new st(o,o.nextSibling,this,t):c.type===1?y=new c.ctor(o,c.name,c.strings,this,t):c.type===6&&(y=new Pt(o,this,t)),this._$AV.push(y),c=i[++f]}a!==c?.index&&(o=K.nextNode(),a++)}return K.currentNode=G,s}p(t){let e=0;for(let i of this._$AV)i!==void 0&&(i.strings!==void 0?(i._$AI(t,i,e),e+=i.strings.length-2):i._$AI(t[e])),e++}},st=class r{get _$AU(){return this._$AM?._$AU??this._$Cv}constructor(t,e,i,s){this.type=2,this._$AH=h,this._$AN=void 0,this._$AA=t,this._$AB=e,this._$AM=i,this.options=s,this._$Cv=s?.isConnected??!0}get parentNode(){let t=this._$AA.parentNode,e=this._$AM;return e!==void 0&&t?.nodeType===11&&(t=e.parentNode),t}get startNode(){return this._$AA}get endNode(){return this._$AB}_$AI(t,e=this){t=Q(this,t,e),dt(t)?t===h||t==null||t===""?(this._$AH!==h&&this._$AR(),this._$AH=h):t!==this._$AH&&t!==R&&this._(t):t._$litType$!==void 0?this.$(t):t.nodeType!==void 0?this.T(t):pe(t)?this.k(t):this._(t)}O(t){return this._$AA.parentNode.insertBefore(t,this._$AB)}T(t){this._$AH!==t&&(this._$AR(),this._$AH=this.O(t))}_(t){this._$AH!==h&&dt(this._$AH)?this._$AA.nextSibling.data=t:this.T(G.createTextNode(t)),this._$AH=t}$(t){let{values:e,_$litType$:i}=t,s=typeof i=="number"?this._$AC(t):(i.el===void 0&&(i.el=pt.createElement(ue(i.h,i.h[0]),this.options)),i);if(this._$AH?._$AD===s)this._$AH.p(e);else{let o=new Et(s,this),a=o.u(this.options);o.p(e),this.T(a),this._$AH=o}}_$AC(t){let e=de.get(t.strings);return e===void 0&&de.set(t.strings,e=new pt(t)),e}k(t){Gt(this._$AH)||(this._$AH=[],this._$AR());let e=this._$AH,i,s=0;for(let o of t)s===e.length?e.push(i=new r(this.O(ct()),this.O(ct()),this,this.options)):i=e[s],i._$AI(o),s++;s<e.length&&(this._$AR(i&&i._$AB.nextSibling,s),e.length=s)}_$AR(t=this._$AA.nextSibling,e){for(this._$AP?.(!1,!0,e);t!==this._$AB;){let i=oe(t).nextSibling;oe(t).remove(),t=i}}setConnected(t){this._$AM===void 0&&(this._$Cv=t,this._$AP?.(t))}},Y=class{get tagName(){return this.element.tagName}get _$AU(){return this._$AM._$AU}constructor(t,e,i,s,o){this.type=1,this._$AH=h,this._$AN=void 0,this.element=t,this.name=e,this._$AM=s,this.options=o,i.length>2||i[0]!==""||i[1]!==""?(this._$AH=Array(i.length-1).fill(new String),this.strings=i):this._$AH=h}_$AI(t,e=this,i,s){let o=this.strings,a=!1;if(o===void 0)t=Q(this,t,e,0),a=!dt(t)||t!==this._$AH&&t!==R,a&&(this._$AH=t);else{let f=t,c,y;for(t=o[0],c=0;c<o.length-1;c++)y=Q(this,f[i+c],e,c),y===R&&(y=this._$AH[c]),a||=!dt(y)||y!==this._$AH[c],y===h?t=h:t!==h&&(t+=(y??"")+o[c+1]),this._$AH[c]=y}a&&!s&&this.j(t)}j(t){t===h?this.element.removeAttribute(this.name):this.element.setAttribute(this.name,t??"")}},St=class extends Y{constructor(){super(...arguments),this.type=3}j(t){this.element[this.name]=t===h?void 0:t}},zt=class extends Y{constructor(){super(...arguments),this.type=4}j(t){this.element.toggleAttribute(this.name,!!t&&t!==h)}},Ct=class extends Y{constructor(t,e,i,s,o){super(t,e,i,s,o),this.type=5}_$AI(t,e=this){if((t=Q(this,t,e,0)??h)===R)return;let i=this._$AH,s=t===h&&i!==h||t.capture!==i.capture||t.once!==i.once||t.passive!==i.passive,o=t!==h&&(i===h||s);s&&this.element.removeEventListener(this.name,this,i),o&&this.element.addEventListener(this.name,this,t),this._$AH=t}handleEvent(t){typeof this._$AH=="function"?this._$AH.call(this.options?.host??this.element,t):this._$AH.handleEvent(t)}},Pt=class{constructor(t,e,i){this.element=t,this.type=6,this._$AN=void 0,this._$AM=e,this.options=i}get _$AU(){return this._$AM._$AU}_$AI(t){Q(this,t)}},fe={M:Wt,P:N,A:Kt,C:1,L:be,R:Et,D:pe,V:Q,I:st,H:Y,N:zt,U:Ct,B:St,F:Pt},Be=Jt.litHtmlPolyfillSupport;Be?.(pt,st),(Jt.litHtmlVersions??=[]).push("3.3.3");var ve=(r,t,e)=>{let i=e?.renderBefore??t,s=i._$litPart$;if(s===void 0){let o=e?.renderBefore??null;i._$litPart$=s=new st(t.insertBefore(ct(),o),o,void 0,e??{})}return s._$AI(r),s};var Yt=globalThis,d=class extends T{constructor(){super(...arguments),this.renderOptions={host:this},this._$Do=void 0}createRenderRoot(){let t=super.createRenderRoot();return this.renderOptions.renderBefore??=t.firstChild,t}update(t){let e=this.render();this.hasUpdated||(this.renderOptions.isConnected=this.isConnected),super.update(t),this._$Do=ve(e,this.renderRoot,this.renderOptions)}connectedCallback(){super.connectedCallback(),this._$Do?.setConnected(!0)}disconnectedCallback(){super.disconnectedCallback(),this._$Do?.setConnected(!1)}render(){return R}};d._$litElement$=!0,d.finalized=!0,Yt.litElementHydrateSupport?.({LitElement:d});var je=Yt.litElementPolyfillSupport;je?.({LitElement:d});(Yt.litElementVersions??=[]).push("4.2.2");var He={attribute:!0,type:String,converter:lt,reflect:!1,hasChanged:wt},Ve=(r=He,t,e)=>{let{kind:i,metadata:s}=e,o=globalThis.litPropertyMetadata.get(s);if(o===void 0&&globalThis.litPropertyMetadata.set(s,o=new Map),i==="setter"&&((r=Object.create(r)).wrapped=!0),o.set(e.name,r),i==="accessor"){let{name:a}=e;return{set(f){let c=t.get.call(this);t.set.call(this,f),this.requestUpdate(a,c,r,!0,f)},init(f){return f!==void 0&&this.C(a,void 0,r,f),f}}}if(i==="setter"){let{name:a}=e;return function(f){let c=this[a];t.call(this,f),this.requestUpdate(a,c,r,!0,f)}}throw Error("Unsupported decorator location: "+i)};function n(r){return(t,e)=>typeof e=="object"?Ve(r,t,e):((i,s,o)=>{let a=s.hasOwnProperty(o);return s.constructor.createProperty(o,i),a?Object.getOwnPropertyDescriptor(s,o):void 0})(r,t,e)}function Z(r){return n({...r,state:!0,attribute:!1})}function w(r,t,e){r.setFormValue(t,t)}function D(r,t,e="",i){r.setValidity(t,e,i)}function U(r){r.setValidity({})}function j(r,t,e="Please fill out this field."){return r?{flags:{customError:!0},message:r}:t?{flags:{valueMissing:!0},message:e}:{flags:{},message:""}}function m(r,t){customElements.get(r)||customElements.define(r,t)}var u=p`
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
`,rt=p`
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
`;function it(r,t,e){return r?{labelText:r,hideVisually:t,controlAriaLabel:""}:{labelText:"",hideVisually:!1,controlAriaLabel:e}}var Ie=Object.defineProperty,O=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Ie(t,e,s),s},S=class extends d{constructor(){super(...arguments),this.variant="primary",this.size="md",this.type="button",this.disabled=!1,this.loading=!1,this.name="",this.value="",this.href="",this.target="",this.rel="",this.iconOnly=!1,this.#t=this.attachInternals(),this.#e=!1}static{this.formAssociated=!0}static{this.styles=[u,p`
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
    `]}#t;#e;get#s(){return this.disabled||this.loading||this.#e}get#r(){return!!this.href}get#o(){return this.getAttribute("aria-label")??""}formDisabledCallback(t){this.#e=t,this.requestUpdate()}#i(t){if(this.#s){t.preventDefault(),t.stopImmediatePropagation();return}if(this.#r)return;let e=this.#t.form;e&&(this.type==="submit"?(this.name&&w(this.#t,this.value),e.requestSubmit(),queueMicrotask(()=>w(this.#t,null))):this.type==="reset"&&e.reset())}render(){let t=l`
      ${this.loading?l`<span class="spinner" aria-hidden="true"></span>`:h}
      <slot></slot>
    `,e=this.#o||h;return this.#r?l`
        <a
          part="base"
          class="base"
          href=${this.#s?h:this.href}
          target=${this.target||h}
          rel=${this.rel||(this.target==="_blank"?"noopener noreferrer":h)}
          aria-disabled=${this.#s?"true":"false"}
          aria-busy=${this.loading?"true":"false"}
          aria-label=${e}
          @click=${this.#i}
        >
          ${t}
        </a>
      `:l`
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
    `}};O([n({reflect:!0})],S.prototype,"variant");O([n({reflect:!0})],S.prototype,"size");O([n({reflect:!0})],S.prototype,"type");O([n({type:Boolean,reflect:!0})],S.prototype,"disabled");O([n({type:Boolean,reflect:!0})],S.prototype,"loading");O([n({reflect:!0})],S.prototype,"name");O([n()],S.prototype,"value");O([n({reflect:!0})],S.prototype,"href");O([n({reflect:!0})],S.prototype,"target");O([n({reflect:!0})],S.prototype,"rel");O([n({type:Boolean,reflect:!0,attribute:"icon-only"})],S.prototype,"iconOnly");m("mb-button",S);var Fe=Object.defineProperty,Je=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Fe(t,e,s),s},Lt=class extends d{constructor(){super(...arguments),this.variant="neutral"}static{this.styles=[u,p`
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
    `]}render(){return l`<span part="base"><slot></slot></span>`}};Je([n({reflect:!0})],Lt.prototype,"variant");m("mb-badge",Lt);var We=Object.defineProperty,Ke=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&We(t,e,s),s},Mt=class extends d{constructor(){super(...arguments),this.variant="info"}static{this.styles=[u,p`
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
    `]}get#t(){return this.variant==="warning"||this.variant==="danger"?"alert":"status"}render(){return l`
      <div part="base" class="alert" role=${this.#t}>
        <slot></slot>
      </div>
    `}};Ke([n({reflect:!0})],Mt.prototype,"variant");m("mb-alert",Mt);var Ge=Object.defineProperty,ye=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Ge(t,e,s),s},mt=class extends d{constructor(){super(...arguments),this._hasHeader=!1,this._hasFooter=!1}static{this.styles=[u,p`
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
    `]}#t(t){let e=t.target;this._hasHeader=e.assignedNodes({flatten:!0}).length>0,this.toggleAttribute("data-has-header",this._hasHeader)}#e(t){let e=t.target;this._hasFooter=e.assignedNodes({flatten:!0}).length>0,this.toggleAttribute("data-has-footer",this._hasFooter)}render(){return l`
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
    `}};ye([Z()],mt.prototype,"_hasHeader");ye([Z()],mt.prototype,"_hasFooter");m("mb-card",mt);var Qe=Object.defineProperty,A=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Qe(t,e,s),s},$=class extends d{constructor(){super(...arguments),this.label="",this.hint="",this.error="",this.value="",this.name="",this.placeholder="",this.type="text",this.disabled=!1,this.required=!1,this.invalid=!1,this.density="default",this.hideLabel=!1,this.min="",this.max="",this.step="",this.accept="",this.multiple=!1,this.#t=this.attachInternals(),this.#e=!1,this.#r="",this.#o=!1,this.#i=!1}static{this.formAssociated=!0}static{this.styles=[u,rt,p`
      :host {
        display: block;
      }

      input[type='file'].control {
        padding-block: var(--mb-space-2);
      }
    `]}#t;#e;#s;#r;#o;#i;get#c(){return this.disabled||this.#e}get#a(){return this.type==="file"}get#n(){return this.getAttribute("aria-label")??""}connectedCallback(){super.connectedCallback(),this.#o||(this.#r=this.value,this.#o=!0)}firstUpdated(){this.#s=this.renderRoot.querySelector("input")??void 0,this.#l()}updated(t){(t.has("value")||t.has("required")||t.has("error")||t.has("disabled")||t.has("name")||t.has("type"))&&this.#l()}formDisabledCallback(t){this.#e=t,this.requestUpdate()}formResetCallback(){this.#i=!1,this.value=this.#r,this.error="",this.invalid=!1,this.#a&&this.#s&&(this.#s.value="")}#h(){let t=this.#s?.files;if(!this.name||!t?.length){w(this.#t,null);return}if(t.length===1){w(this.#t,t[0]);return}let e=new FormData;for(let i of t)e.append(this.name,i);w(this.#t,e)}#l(){this.#a?this.#h():w(this.#t,this.name?this.value:null);let t=this.required&&(this.#a?!this.#s?.files?.length:!this.value),{flags:e,message:i}=j(this.error,t);i?(D(this.#t,e,i,this.#s),this.invalid=!!this.error||this.#i):(U(this.#t),this.invalid=!1)}#d(t){let e=t.target;this.#i=!0,this.#a||(this.value=e.value),this.#l(),this.dispatchEvent(new CustomEvent("mb-input",{detail:{value:this.value,files:e.files},bubbles:!0,composed:!0}))}#p(t){let e=t.target;this.#i=!0,this.#a||(this.value=e.value),this.#l(),this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value,files:e.files},bubbles:!0,composed:!0}))}#m(t){if(t.key!=="Enter"||t.defaultPrevented||this.#a)return;let e=this.#t.form;e&&(t.preventDefault(),e.requestSubmit())}render(){let t=[this.hint&&!this.error?"hint":"",this.error?"error":""].filter(Boolean).join(" "),{labelText:e,hideVisually:i,controlAriaLabel:s}=it(this.label,this.hideLabel,this.#n);return l`
      <div class="field">
        ${e?l`<label
              part="label"
              class="label${i?" visually-hidden":""}"
              for="control"
              >${e}</label
            >`:h}
        <input
          id="control"
          part="control"
          class="control"
          .type=${this.type}
          .value=${this.#a?"":this.value}
          name=${this.name||h}
          placeholder=${this.placeholder||h}
          min=${this.type==="number"&&this.min!==""?this.min:h}
          max=${this.type==="number"&&this.max!==""?this.max:h}
          step=${this.type==="number"&&this.step!==""?this.step:h}
          accept=${this.#a&&this.accept?this.accept:h}
          ?multiple=${this.#a&&this.multiple}
          ?disabled=${this.#c}
          ?required=${this.required}
          aria-invalid=${this.invalid?"true":"false"}
          aria-label=${s||h}
          aria-describedby=${t||h}
          @input=${this.#d}
          @change=${this.#p}
          @keydown=${this.#m}
        />
        ${this.hint&&!this.error?l`<p id="hint" class="hint">${this.hint}</p>`:h}
        ${this.error?l`<p id="error" class="error" role="alert">${this.error}</p>`:h}
      </div>
    `}};A([n()],$.prototype,"label");A([n()],$.prototype,"hint");A([n()],$.prototype,"error");A([n()],$.prototype,"value");A([n({reflect:!0})],$.prototype,"name");A([n()],$.prototype,"placeholder");A([n({reflect:!0})],$.prototype,"type");A([n({type:Boolean,reflect:!0})],$.prototype,"disabled");A([n({type:Boolean,reflect:!0})],$.prototype,"required");A([n({type:Boolean,reflect:!0})],$.prototype,"invalid");A([n({reflect:!0})],$.prototype,"density");A([n({type:Boolean,reflect:!0,attribute:"hide-label"})],$.prototype,"hideLabel");A([n()],$.prototype,"min");A([n()],$.prototype,"max");A([n()],$.prototype,"step");A([n()],$.prototype,"accept");A([n({type:Boolean})],$.prototype,"multiple");m("mb-input",$);var Ye=Object.defineProperty,C=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Ye(t,e,s),s},E=class extends d{constructor(){super(...arguments),this.label="",this.hint="",this.error="",this.value="",this.name="",this.placeholder="",this.disabled=!1,this.required=!1,this.invalid=!1,this.rows=4,this.density="default",this.hideLabel=!1,this.#t=this.attachInternals(),this.#e=!1,this.#r="",this.#o=!1,this.#i=!1}static{this.formAssociated=!0}static{this.styles=[u,rt,p`
      :host {
        display: block;
      }

      textarea.control {
        min-block-size: 6rem;
        resize: vertical;
      }
    `]}#t;#e;#s;#r;#o;#i;get#c(){return this.disabled||this.#e}get#a(){return this.getAttribute("aria-label")??""}connectedCallback(){super.connectedCallback(),this.#o||(this.#r=this.value,this.#o=!0)}firstUpdated(){this.#s=this.renderRoot.querySelector("textarea")??void 0,this.#n()}updated(t){(t.has("value")||t.has("required")||t.has("error")||t.has("disabled")||t.has("name"))&&this.#n()}formDisabledCallback(t){this.#e=t,this.requestUpdate()}formResetCallback(){this.#i=!1,this.value=this.#r,this.error="",this.invalid=!1}#n(){w(this.#t,this.name?this.value:null);let t=this.required&&!this.value,{flags:e,message:i}=j(this.error,t);i?(D(this.#t,e,i,this.#s),this.invalid=!!this.error||this.#i):(U(this.#t),this.invalid=!1)}#h(t){let e=t.target;this.#i=!0,this.value=e.value,this.dispatchEvent(new CustomEvent("mb-input",{detail:{value:this.value},bubbles:!0,composed:!0}))}#l(t){let e=t.target;this.#i=!0,this.value=e.value,this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value},bubbles:!0,composed:!0}))}render(){let t=[this.hint&&!this.error?"hint":"",this.error?"error":""].filter(Boolean).join(" "),{labelText:e,hideVisually:i,controlAriaLabel:s}=it(this.label,this.hideLabel,this.#a);return l`
      <div class="field">
        ${e?l`<label
              part="label"
              class="label${i?" visually-hidden":""}"
              for="control"
              >${e}</label
            >`:h}
        <textarea
          id="control"
          part="control"
          class="control"
          .value=${this.value}
          name=${this.name||h}
          placeholder=${this.placeholder||h}
          rows=${this.rows}
          ?disabled=${this.#c}
          ?required=${this.required}
          aria-invalid=${this.invalid?"true":"false"}
          aria-label=${s||h}
          aria-describedby=${t||h}
          @input=${this.#h}
          @change=${this.#l}
        ></textarea>
        ${this.hint&&!this.error?l`<p id="hint" class="hint">${this.hint}</p>`:h}
        ${this.error?l`<p id="error" class="error" role="alert">${this.error}</p>`:h}
      </div>
    `}};C([n()],E.prototype,"label");C([n()],E.prototype,"hint");C([n()],E.prototype,"error");C([n()],E.prototype,"value");C([n({reflect:!0})],E.prototype,"name");C([n()],E.prototype,"placeholder");C([n({type:Boolean,reflect:!0})],E.prototype,"disabled");C([n({type:Boolean,reflect:!0})],E.prototype,"required");C([n({type:Boolean,reflect:!0})],E.prototype,"invalid");C([n({type:Number})],E.prototype,"rows");C([n({reflect:!0})],E.prototype,"density");C([n({type:Boolean,reflect:!0,attribute:"hide-label"})],E.prototype,"hideLabel");m("mb-textarea",E);var Ze=Object.defineProperty,B=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Ze(t,e,s),s},P=class extends d{constructor(){super(...arguments),this.label="",this.error="",this.name="",this.value="on",this.checked=!1,this.indeterminate=!1,this.disabled=!1,this.required=!1,this.invalid=!1,this.#t=this.attachInternals(),this.#e=!1,this.#r=!1,this.#o=!1,this.#i=!1}static{this.formAssociated=!0}static{this.styles=[u,p`
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
    `]}#t;#e;#s;#r;#o;#i;get#c(){return this.disabled||this.#e}connectedCallback(){super.connectedCallback(),this.#o||(this.#r=this.checked,this.#o=!0)}firstUpdated(){this.#s=this.renderRoot.querySelector("input")??void 0,this.#a(),this.#n()}updated(t){t.has("indeterminate")&&this.#a(),(t.has("checked")||t.has("value")||t.has("required")||t.has("error")||t.has("disabled")||t.has("name"))&&this.#n()}formDisabledCallback(t){this.#e=t,this.requestUpdate()}formResetCallback(){this.#i=!1,this.checked=this.#r,this.indeterminate=!1,this.error="",this.invalid=!1}#a(){this.#s&&(this.#s.indeterminate=this.indeterminate)}#n(){w(this.#t,this.name&&this.checked?this.value:null);let t=this.required&&!this.checked,e=this.error||(t?"Please check this box.":"");if(e){let i=this.error?{customError:!0}:{valueMissing:!0};D(this.#t,i,e,this.#s),this.invalid=!!this.error||this.#i}else U(this.#t),this.invalid=!1}#h(t){let e=t.target;this.#i=!0,this.checked=e.checked,this.indeterminate=!1,this.dispatchEvent(new CustomEvent("mb-change",{detail:{checked:this.checked,value:this.value},bubbles:!0,composed:!0}))}render(){let t=this.error?"error":"";return l`
      <label part="label">
        <input
          part="control"
          type="checkbox"
          .checked=${this.checked}
          name=${this.name||h}
          value=${this.value}
          ?disabled=${this.#c}
          ?required=${this.required}
          aria-invalid=${this.invalid?"true":"false"}
          aria-describedby=${t||h}
          @change=${this.#h}
        />
        <span>${this.label}<slot></slot></span>
      </label>
      ${this.error?l`<p id="error" class="error" role="alert">${this.error}</p>`:h}
    `}};B([n()],P.prototype,"label");B([n()],P.prototype,"error");B([n({reflect:!0})],P.prototype,"name");B([n()],P.prototype,"value");B([n({type:Boolean,reflect:!0})],P.prototype,"checked");B([n({type:Boolean,reflect:!0})],P.prototype,"indeterminate");B([n({type:Boolean,reflect:!0})],P.prototype,"disabled");B([n({type:Boolean,reflect:!0})],P.prototype,"required");B([n({type:Boolean,reflect:!0})],P.prototype,"invalid");m("mb-checkbox",P);var ge={ATTRIBUTE:1,CHILD:2,PROPERTY:3,BOOLEAN_ATTRIBUTE:4,EVENT:5,ELEMENT:6},$e=r=>(...t)=>({_$litDirective$:r,values:t}),Dt=class{constructor(t){}get _$AU(){return this._$AM._$AU}_$AT(t,e,i){this._$Ct=t,this._$AM=e,this._$Ci=i}_$AS(t,e){return this.update(t,e)}update(t,e){return this.render(...e)}};var{I:Xe}=fe,xe=r=>r;var _e=()=>document.createComment(""),ot=(r,t,e)=>{let i=r._$AA.parentNode,s=t===void 0?r._$AB:t._$AA;if(e===void 0){let o=i.insertBefore(_e(),s),a=i.insertBefore(_e(),s);e=new Xe(o,a,r,r.options)}else{let o=e._$AB.nextSibling,a=e._$AM,f=a!==r;if(f){let c;e._$AQ?.(r),e._$AM=r,e._$AP!==void 0&&(c=r._$AU)!==a._$AU&&e._$AP(c)}if(o!==s||f){let c=e._$AA;for(;c!==o;){let y=xe(c).nextSibling;xe(i).insertBefore(c,s),c=y}}}return e},H=(r,t,e=r)=>(r._$AI(t,e),r),ts={},Ae=(r,t=ts)=>r._$AH=t,we=r=>r._$AH,Ut=r=>{r._$AR(),r._$AA.remove()};var ke=(r,t,e)=>{let i=new Map;for(let s=t;s<=e;s++)i.set(r[s],s);return i},qt=$e(class extends Dt{constructor(r){if(super(r),r.type!==ge.CHILD)throw Error("repeat() can only be used in text expressions")}dt(r,t,e){let i;e===void 0?e=t:t!==void 0&&(i=t);let s=[],o=[],a=0;for(let f of r)s[a]=i?i(f,a):a,o[a]=e(f,a),a++;return{values:o,keys:s}}render(r,t,e){return this.dt(r,t,e).values}update(r,[t,e,i]){let s=we(r),{values:o,keys:a}=this.dt(t,e,i);if(!Array.isArray(s))return this.ut=a,o;let f=this.ut??=[],c=[],y,x,b=0,g=s.length-1,v=0,_=o.length-1;for(;b<=g&&v<=_;)if(s[b]===null)b++;else if(s[g]===null)g--;else if(f[b]===a[v])c[v]=H(s[b],o[v]),b++,v++;else if(f[g]===a[_])c[_]=H(s[g],o[_]),g--,_--;else if(f[b]===a[_])c[_]=H(s[b],o[_]),ot(r,c[_+1],s[b]),b++,_--;else if(f[g]===a[v])c[v]=H(s[g],o[v]),ot(r,s[b],s[g]),g--,v++;else if(y===void 0&&(y=ke(a,v,_),x=ke(f,b,g)),y.has(f[b]))if(y.has(f[g])){let q=x.get(a[v]),Ht=q!==void 0?s[q]:null;if(Ht===null){let Xt=ot(r,s[b]);H(Xt,o[v]),c[v]=Xt}else c[v]=H(Ht,o[v]),ot(r,s[b],Ht),s[q]=null;v++}else Ut(s[g]),g--;else Ut(s[b]),b++;for(;v<=_;){let q=ot(r,c[_+1]);H(q,o[v]),c[v++]=q}for(;b<=g;){let q=s[b++];q!==null&&Ut(q)}return this.ut=a,Ae(r,c),R}});var es=Object.defineProperty,z=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&es(t,e,s),s};function ss(r){if(!r)return[];try{let t=JSON.parse(r);return Array.isArray(t)?t.filter(e=>!!e&&typeof e=="object"&&typeof e.value=="string"&&typeof e.label=="string").map(e=>({value:e.value,label:e.label,disabled:!!e.disabled})):[]}catch{return[]}}var k=class extends d{constructor(){super(...arguments),this.label="",this.hint="",this.error="",this.value="",this.name="",this.disabled=!1,this.required=!1,this.invalid=!1,this.density="default",this.hideLabel=!1,this.placeholder="",this.options=[],this._slottedOptions=[],this.#t=this.attachInternals(),this.#e=!1,this.#r="",this.#o=!1,this.#i=!1}static{this.formAssociated=!0}static{this.styles=[u,rt,p`
      :host {
        display: block;
      }

      slot[name='options'] {
        display: none;
      }
    `]}#t;#e;#s;#r;#o;#i;get#c(){return this.disabled||this.#e}get#a(){return this._slottedOptions.length?this._slottedOptions:this.options}get#n(){return this.#a.filter(t=>t.value!=="")}get#h(){let t=this.#a.find(e=>e.value==="");return t?.label?t.label:this.placeholder}get#l(){return this.getAttribute("aria-label")??""}connectedCallback(){super.connectedCallback(),this.#o||(this.#r=this.value,this.#o=!0),this.#p()}firstUpdated(){this.#s=this.renderRoot.querySelector("select")??void 0,this.#b()}updated(t){(t.has("value")||t.has("required")||t.has("error")||t.has("options")||t.has("_slottedOptions")||t.has("disabled")||t.has("name"))&&this.#b()}formDisabledCallback(t){this.#e=t,this.requestUpdate()}formResetCallback(){this.#i=!1,this.value=this.#r,this.error="",this.invalid=!1}#d(t){return t instanceof HTMLOptionElement?{value:t.value,label:t.label||t.textContent?.trim()||t.value,disabled:t.disabled}:null}#p(){let t=[...this.querySelectorAll(":scope > option")].map(e=>this.#d(e)).filter(e=>e!=null);t.length&&(this._slottedOptions=t)}#m(){let t=this.renderRoot.querySelector('slot[name="options"]'),e=this.renderRoot.querySelector("slot:not([name])"),i=[...t?.assignedElements({flatten:!0})??[],...e?.assignedElements({flatten:!0})??[]].map(a=>this.#d(a)).filter(a=>a!=null),s=JSON.stringify(this._slottedOptions),o=JSON.stringify(i);s!==o&&(this._slottedOptions=i)}#u(){this.#m()}#b(){this.#s&&this.#s.value!==this.value&&(this.#s.value=this.value),w(this.#t,this.name?this.value:null);let t=this.required&&!this.value,{flags:e,message:i}=j(this.error,t,"Please select an option.");i?(D(this.#t,e,i,this.#s),this.invalid=!!this.error||this.#i):(U(this.#t),this.invalid=!1)}#f(t){let e=t.target;this.#i=!0,this.value=e.value,this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value},bubbles:!0,composed:!0}))}render(){let t=[this.hint&&!this.error?"hint":"",this.error?"error":""].filter(Boolean).join(" "),{labelText:e,hideVisually:i,controlAriaLabel:s}=it(this.label,this.hideLabel,this.#l);return l`
      <div class="field">
        ${e?l`<label
              part="label"
              class="label${i?" visually-hidden":""}"
              for="control"
              >${e}</label
            >`:h}
        <select
          id="control"
          part="control"
          class="control"
          name=${this.name||h}
          ?disabled=${this.#c}
          ?required=${this.required}
          aria-invalid=${this.invalid?"true":"false"}
          aria-label=${s||h}
          aria-describedby=${t||h}
          .value=${this.value}
          @change=${this.#f}
        >
          <option value="" ?disabled=${this.required}>${this.#h}</option>
          ${qt(this.#n,o=>o.value,o=>l`
              <option value=${o.value} ?disabled=${!!o.disabled}>
                ${o.label}
              </option>
            `)}
        </select>
        ${this.hint&&!this.error?l`<p id="hint" class="hint">${this.hint}</p>`:h}
        ${this.error?l`<p id="error" class="error" role="alert">${this.error}</p>`:h}
      </div>
      <slot name="options" @slotchange=${this.#u}></slot>
      <slot @slotchange=${this.#u}></slot>
    `}};z([n()],k.prototype,"label");z([n()],k.prototype,"hint");z([n()],k.prototype,"error");z([n()],k.prototype,"value");z([n({reflect:!0})],k.prototype,"name");z([n({type:Boolean,reflect:!0})],k.prototype,"disabled");z([n({type:Boolean,reflect:!0})],k.prototype,"required");z([n({type:Boolean,reflect:!0})],k.prototype,"invalid");z([n({reflect:!0})],k.prototype,"density");z([n({type:Boolean,reflect:!0,attribute:"hide-label"})],k.prototype,"hideLabel");z([n()],k.prototype,"placeholder");z([n({attribute:"options",converter:{fromAttribute:ss,toAttribute(r){return r?.length?JSON.stringify(r):null}}})],k.prototype,"options");z([Z()],k.prototype,"_slottedOptions");m("mb-select",k);var rs=Object.defineProperty,Ee=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&rs(t,e,s),s},ut=class extends d{constructor(){super(...arguments),this.open=!1,this.heading="",this.#e=!1}static{this.styles=[u,p`
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
    `]}#t;#e;firstUpdated(){this.#t=this.renderRoot.querySelector("dialog")??void 0,this.#t?.addEventListener("close",()=>{this.#e||(this.open&&(this.open=!1),this.#r())}),this.#s()}updated(t){t.has("open")&&this.#s()}#s(){let t=this.#t;t&&(this.open&&!t.open?t.showModal():!this.open&&t.open&&(this.#e=!0,t.close(),this.#e=!1,this.#r()))}#r(){this.dispatchEvent(new CustomEvent("mb-close",{bubbles:!0,composed:!0}))}close(){!this.open&&!this.#t?.open||(this.open=!1)}#o(){this.close()}render(){return l`
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
    `}};Ee([n({type:Boolean,reflect:!0})],ut.prototype,"open");Ee([n()],ut.prototype,"heading");m("mb-modal",ut);var is=Object.defineProperty,Tt=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&is(t,e,s),s},X=class extends d{constructor(){super(...arguments),this.value=0,this.max=100,this.percent=null,this.label=""}static{this.styles=[u,p`
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
    `]}get#t(){if(this.percent!=null&&!Number.isNaN(this.percent))return Math.min(100,Math.max(0,this.percent));let t=this.max>0?this.max:100;return Math.min(100,Math.max(0,this.value/t*100))}get#e(){return this.percent!=null&&!Number.isNaN(this.percent)?this.#t:this.value}get#s(){return this.percent!=null&&!Number.isNaN(this.percent)?100:this.max>0?this.max:100}render(){let t=this.#t;return l`
      <div class="wrap">
        ${this.label?l`<div part="label" class="label" id="label">${this.label}</div>`:h}
        <div
          part="track"
          class="track"
          role="progressbar"
          aria-valuemin="0"
          aria-valuenow=${this.#e}
          aria-valuemax=${this.#s}
          aria-labelledby=${this.label?"label":h}
          aria-label=${this.label?h:this.getAttribute("aria-label")||"Progress"}
        >
          <div part="bar" class="bar" style="inline-size: ${t}%"></div>
        </div>
        <slot></slot>
      </div>
    `}};Tt([n({type:Number})],X.prototype,"value");Tt([n({type:Number})],X.prototype,"max");Tt([n({type:Number})],X.prototype,"percent");Tt([n()],X.prototype,"label");m("mb-progress",X);var os=Object.defineProperty,as=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&os(t,e,s),s},Nt=class extends d{constructor(){super(...arguments),this.label="Filters"}static{this.styles=[u,p`
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
    `]}render(){return l`
      <nav part="nav" class="scroller" aria-label=${this.label}>
        <div part="list" class="list" role="list">
          <slot></slot>
        </div>
      </nav>
    `}};as([n()],Nt.prototype,"label");m("mb-segmented-control",Nt);var ns=Object.defineProperty,ls=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&ns(t,e,s),s},Rt=class extends d{constructor(){super(...arguments),this.heading=""}static{this.styles=[u,p`
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
    `]}#t(t){let e=t.target.assignedNodes({flatten:!0}).length>0;this.renderRoot.querySelector(".actions")?.toggleAttribute("data-has-content",e)}render(){return l`
      <div part="panel" class="panel">
        ${this.heading?l`<h2 part="heading" class="heading">${this.heading}</h2>`:l`<slot name="heading"></slot>`}
        <div part="body" class="body">
          <slot></slot>
        </div>
        <div part="actions" class="actions">
          <slot name="actions" @slotchange=${this.#t}></slot>
        </div>
      </div>
    `}};ls([n()],Rt.prototype,"heading");m("mb-empty-state",Rt);var hs=Object.defineProperty,V=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&hs(t,e,s),s},L=class extends d{constructor(){super(...arguments),this.prevUrl="",this.nextUrl="",this.prevDisabled=!1,this.nextDisabled=!1,this.status="",this.prevLabel="Previous",this.nextLabel="Next",this.label="Pagination"}static{this.styles=[u,p`
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
    `]}render(){let t=this.prevDisabled||!this.prevUrl,e=this.nextDisabled||!this.nextUrl;return l`
      <nav part="nav" aria-label=${this.label}>
        <div part="status" class="status">${this.status}<slot name="status"></slot></div>
        <div part="actions" class="actions">
          <slot name="prev">
            ${t?l`<span class="disabled" aria-disabled="true">${this.prevLabel}</span>`:l`<a part="prev" href=${this.prevUrl}>${this.prevLabel}</a>`}
          </slot>
          <slot name="next">
            ${e?l`<span class="disabled" aria-disabled="true">${this.nextLabel}</span>`:l`<a part="next" href=${this.nextUrl}>${this.nextLabel}</a>`}
          </slot>
        </div>
      </nav>
    `}};V([n({attribute:"prev-url"})],L.prototype,"prevUrl");V([n({attribute:"next-url"})],L.prototype,"nextUrl");V([n({type:Boolean,attribute:"prev-disabled"})],L.prototype,"prevDisabled");V([n({type:Boolean,attribute:"next-disabled"})],L.prototype,"nextDisabled");V([n()],L.prototype,"status");V([n({attribute:"prev-label"})],L.prototype,"prevLabel");V([n({attribute:"next-label"})],L.prototype,"nextLabel");V([n()],L.prototype,"label");m("mb-pagination",L);var cs=Object.defineProperty,Bt=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&cs(t,e,s),s},tt=class extends d{constructor(){super(...arguments),this.open=!1,this.variant="info",this.autoDismiss=4e3,this.message="",this.#t=0,this.#e=t=>{let e=t.detail;e&&(e.variant&&(this.variant=e.variant),e.message!=null&&(this.message=e.message),e.autoDismiss!=null&&(this.autoDismiss=e.autoDismiss),this.show())}}static{this.styles=[u,p`
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
    `]}#t;#e;connectedCallback(){super.connectedCallback(),document.addEventListener("mb-toast",this.#e)}disconnectedCallback(){super.disconnectedCallback(),document.removeEventListener("mb-toast",this.#e),this.#r()}updated(t){t.has("open")&&(this.open?this.#s():this.#r())}show(t,e){t!=null&&(this.message=t),e&&(this.variant=e),this.open=!0}hide(){this.open=!1}#s(){this.#r(),this.autoDismiss>0&&(this.#t=window.setTimeout(()=>this.hide(),this.autoDismiss))}#r(){this.#t&&(window.clearTimeout(this.#t),this.#t=0)}#o(){this.hide(),this.dispatchEvent(new CustomEvent("mb-close",{bubbles:!0,composed:!0}))}render(){let t=this.variant==="danger"?"alert":"status";return l`
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
    `}};Bt([n({type:Boolean,reflect:!0})],tt.prototype,"open");Bt([n({reflect:!0})],tt.prototype,"variant");Bt([n({type:Number,attribute:"auto-dismiss"})],tt.prototype,"autoDismiss");Bt([n()],tt.prototype,"message");m("mb-toast",tt);var ds=Object.defineProperty,bt=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&ds(t,e,s),s},I=class extends d{constructor(){super(...arguments),this.value="",this.label="",this.disabled=!1,this.checked=!1,this.name=""}static{this.styles=[u,p`
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
    `]}#t;firstUpdated(){this.#t=this.renderRoot.querySelector("input")??void 0}focus(t){this.#t?.focus(t)}#e(){this.checked=!0,this.dispatchEvent(new CustomEvent("mb-radio-select",{detail:{value:this.value},bubbles:!0,composed:!0}))}render(){return l`
      <label part="label">
        <input
          part="control"
          type="radio"
          name=${this.name||h}
          .value=${this.value}
          .checked=${this.checked}
          ?disabled=${this.disabled}
          @change=${this.#e}
        />
        <span>${this.label}<slot></slot></span>
      </label>
    `}};bt([n()],I.prototype,"value");bt([n()],I.prototype,"label");bt([n({type:Boolean,reflect:!0})],I.prototype,"disabled");bt([n({type:Boolean,reflect:!0})],I.prototype,"checked");bt([n()],I.prototype,"name");m("mb-radio",I);var ps=Object.defineProperty,F=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&ps(t,e,s),s};function ms(r){if(!r)return[];try{let t=JSON.parse(r);return Array.isArray(t)?t.filter(e=>!!e&&typeof e=="object"&&typeof e.value=="string"&&typeof e.label=="string").map(e=>({value:e.value,label:e.label,disabled:!!e.disabled})):[]}catch{return[]}}var M=class extends d{constructor(){super(...arguments),this.label="",this.error="",this.value="",this.name="",this.disabled=!1,this.required=!1,this.invalid=!1,this.options=[],this.#t=this.attachInternals(),this.#e=!1,this.#s="",this.#r=!1,this.#o=!1,this.#h=t=>{let e=t.detail?.value;e!=null&&(this.#o=!0,this.value=e,this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value},bubbles:!0,composed:!0})))},this.#l=t=>{if(!["ArrowDown","ArrowUp","ArrowRight","ArrowLeft"].includes(t.key))return;let e=this.#c().filter(a=>!a.disabled);if(!e.length)return;t.preventDefault();let i=e.findIndex(a=>a.value===this.value),s=t.key==="ArrowDown"||t.key==="ArrowRight"?1:-1,o=e[(i+s+e.length)%e.length];this.#o=!0,this.value=o.value,o.focus(),this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value},bubbles:!0,composed:!0}))}}static{this.formAssociated=!0}static{this.styles=[u,p`
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
    `]}#t;#e;#s;#r;#o;get#i(){return this.disabled||this.#e}connectedCallback(){super.connectedCallback(),this.#r||(this.#s=this.value,this.#r=!0),this.addEventListener("mb-radio-select",this.#h),this.addEventListener("keydown",this.#l)}disconnectedCallback(){super.disconnectedCallback(),this.removeEventListener("mb-radio-select",this.#h),this.removeEventListener("keydown",this.#l)}firstUpdated(){this.#a(),this.#n()}updated(t){(t.has("value")||t.has("name")||t.has("disabled")||t.has("options"))&&this.#a(),(t.has("value")||t.has("required")||t.has("error")||t.has("name")||t.has("disabled"))&&this.#n()}formDisabledCallback(t){this.#e=t,this.requestUpdate(),this.#a()}formResetCallback(){this.#o=!1,this.value=this.#s,this.error="",this.invalid=!1}#c(){let t=this.renderRoot.querySelector("slot")?.assignedElements({flatten:!0}).filter(i=>i.localName==="mb-radio")??[],e=[...this.renderRoot.querySelectorAll(".options > mb-radio")];return[...t,...e]}#a(){let t=this.#c();for(let e of t)e.name=this.name||"mb-radio-group",e.checked=e.value===this.value,this.#i&&(e.disabled=!0)}#n(){w(this.#t,this.name?this.value:null);let t=this.required&&!this.value,{flags:e,message:i}=j(this.error,t,"Please select an option.");i?(D(this.#t,e,i),this.invalid=!!this.error||this.#o):(U(this.#t),this.invalid=!1)}#h;#l;#d(){this.#a()}render(){return l`
      <fieldset part="fieldset" ?disabled=${this.#i}>
        ${this.label?l`<legend part="legend">${this.label}</legend>`:h}
        <div class="options" part="options" role="radiogroup" aria-invalid=${this.invalid?"true":"false"}>
          <slot @slotchange=${this.#d}></slot>
          ${this.options.map(t=>l`
              <mb-radio
                .value=${t.value}
                .label=${t.label}
                ?disabled=${!!t.disabled||this.#i}
                ?checked=${t.value===this.value}
                .name=${this.name||"mb-radio-group"}
              ></mb-radio>
            `)}
        </div>
        ${this.error?l`<p class="error" role="alert">${this.error}</p>`:h}
      </fieldset>
    `}};F([n()],M.prototype,"label");F([n()],M.prototype,"error");F([n()],M.prototype,"value");F([n({reflect:!0})],M.prototype,"name");F([n({type:Boolean,reflect:!0})],M.prototype,"disabled");F([n({type:Boolean,reflect:!0})],M.prototype,"required");F([n({type:Boolean,reflect:!0})],M.prototype,"invalid");F([n({attribute:"options",converter:{fromAttribute:ms,toAttribute(r){return r?.length?JSON.stringify(r):null}}})],M.prototype,"options");m("mb-radio-group",M);var us=Object.defineProperty,Se=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&us(t,e,s),s},ft=class extends d{constructor(){super(...arguments),this.href="",this.size="md"}static{this.styles=[u,p`
      :host {
        display: inline-flex;
        max-inline-size: 100%;
      }

      .tag {
        display: inline-flex;
        align-items: center;
        gap: var(--mb-space-1);
        max-inline-size: 100%;
        padding-block: 0.15rem;
        padding-inline: var(--mb-space-2);
        border: 1px solid var(--mb-color-border);
        border-radius: var(--mb-radius-sm);
        background: var(--mb-color-surface);
        color: var(--mb-color-fg);
        font-size: var(--mb-font-size-sm);
        font-weight: 600;
        line-height: 1.3;
        text-decoration: none;
        overflow-wrap: anywhere;
      }

      :host([size='sm']) .tag {
        font-size: 0.75rem;
        padding-inline: 0.4rem;
      }
    `]}render(){return this.href?l`
        <a part="base" class="tag" href=${this.href}>
          <slot></slot>
        </a>
      `:l`<span part="base" class="tag"><slot></slot></span>`}};Se([n({reflect:!0})],ft.prototype,"href");Se([n({reflect:!0})],ft.prototype,"size");m("mb-tag",ft);var bs=Object.defineProperty,ze=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&bs(t,e,s),s};function fs(r){if(!r)return[];try{let t=JSON.parse(r);return Array.isArray(t)?t.filter(e=>!!e&&typeof e=="object"&&typeof e.label=="string").map(e=>({label:e.label,href:e.href,current:!!e.current})):[]}catch{return[]}}var vt=class extends d{constructor(){super(...arguments),this.label="Breadcrumb",this.items=[]}static{this.styles=[u,p`
      :host {
        display: block;
      }

      nav ol {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: var(--mb-space-1) var(--mb-space-2);
        margin: 0;
        padding: 0;
        list-style: none;
        font-size: var(--mb-font-size-sm);
      }

      li,
      ::slotted(li) {
        display: inline-flex;
        align-items: center;
        gap: var(--mb-space-2);
        min-inline-size: 0;
        list-style: none;
      }

      li:not(:last-child)::after {
        content: '/';
        color: var(--mb-color-muted);
      }

      a {
        color: var(--mb-color-accent);
        text-decoration: none;
        overflow-wrap: anywhere;
      }

      a:hover {
        text-decoration: underline;
      }

      [aria-current='page'] {
        color: var(--mb-color-muted);
        font-weight: 600;
      }
    `]}render(){return l`
      <nav part="nav" aria-label=${this.label}>
        <ol part="list">
          ${this.items.length?qt(this.items,t=>`${t.href??""}:${t.label}`,t=>l`
                  <li part="item">
                    ${t.current||!t.href?l`<span aria-current=${t.current?"page":h}
                          >${t.label}</span
                        >`:l`<a href=${t.href}>${t.label}</a>`}
                  </li>
                `):l`<slot></slot>`}
        </ol>
      </nav>
    `}};ze([n()],vt.prototype,"label");ze([n({attribute:"items",converter:{fromAttribute:fs,toAttribute(r){return r?.length?JSON.stringify(r):null}}})],vt.prototype,"items");m("mb-breadcrumbs",vt);var vs=Object.defineProperty,Ce=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&vs(t,e,s),s},yt=class extends d{constructor(){super(...arguments),this.label="Primary",this.open=!1}static{this.styles=[u,p`
      :host {
        display: block;
      }

      :host([hidden]) {
        display: none !important;
      }

      nav {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: var(--mb-space-1) var(--mb-space-3);
      }

      ::slotted(a) {
        color: var(--mb-color-muted);
        font-weight: 600;
        font-size: var(--mb-font-size-sm);
        text-decoration: none;
        padding-block: var(--mb-space-2);
        padding-inline: var(--mb-space-2);
        border-radius: var(--mb-radius-sm);
      }

      ::slotted(a:hover) {
        color: var(--mb-color-fg);
      }

      ::slotted(a[aria-current='page']),
      ::slotted(a.is-active) {
        color: var(--mb-color-accent);
        background: var(--mb-color-accent-soft);
      }

      ::slotted(a:focus-visible) {
        outline: var(--mb-focus-ring);
        outline-offset: var(--mb-focus-offset);
      }

      @media (max-width: 36rem) {
        :host(:not([open]):not([data-always-visible])) {
          display: none;
        }
      }
    `]}render(){return l`
      <nav part="nav" aria-label=${this.label}>
        <slot></slot>
      </nav>
    `}};Ce([n()],yt.prototype,"label");Ce([n({type:Boolean,reflect:!0})],yt.prototype,"open");m("mb-nav",yt);var ys=Object.defineProperty,jt=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&ys(t,e,s),s},et=class extends d{constructor(){super(...arguments),this.expanded=!1,this.for="",this.labelOpen="Menu",this.labelClose="Close menu"}static{this.styles=[u,p`
      :host {
        display: none;
      }

      @media (max-width: 36rem) {
        :host {
          display: inline-flex;
        }
      }

      button {
        appearance: none;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        min-inline-size: 2.5rem;
        min-block-size: 2.5rem;
        border: 1px solid var(--mb-color-border);
        border-radius: var(--mb-radius-md);
        background: var(--mb-color-surface);
        color: var(--mb-color-fg);
        cursor: pointer;
        font: inherit;
        font-weight: 700;
      }
    `]}#t(){this.expanded=!this.expanded;let t=this.for?document.getElementById(this.for):null;t&&(t.toggleAttribute("open",this.expanded),"open"in t&&(t.open=this.expanded)),this.dispatchEvent(new CustomEvent("mb-toggle",{detail:{expanded:this.expanded},bubbles:!0,composed:!0}))}render(){return l`
      <button
        part="button"
        type="button"
        aria-expanded=${this.expanded?"true":"false"}
        aria-controls=${this.for||h}
        aria-label=${this.expanded?this.labelClose:this.labelOpen}
        @click=${this.#t}
      >
        <slot>${this.expanded?"\u2715":"\u2630"}</slot>
      </button>
    `}};jt([n({type:Boolean,reflect:!0})],et.prototype,"expanded");jt([n({attribute:"for"})],et.prototype,"for");jt([n({attribute:"label-open"})],et.prototype,"labelOpen");jt([n({attribute:"label-close"})],et.prototype,"labelClose");m("mb-nav-toggle",et);var gs=Object.defineProperty,gt=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&gs(t,e,s),s},J=class extends d{constructor(){super(...arguments),this.src="",this.alt="",this.name="",this.size="md",this._failed=!1}static{this.styles=[u,p`
      :host {
        display: inline-flex;
        vertical-align: middle;
      }

      .avatar {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        overflow: clip;
        border-radius: 50%;
        background: var(--mb-color-accent-soft);
        color: var(--mb-color-accent);
        font-weight: 700;
        line-height: 1;
        user-select: none;
      }

      :host([size='sm']) .avatar {
        inline-size: 1.75rem;
        block-size: 1.75rem;
        font-size: 0.7rem;
      }

      :host([size='md']) .avatar,
      :host(:not([size])) .avatar {
        inline-size: 2.25rem;
        block-size: 2.25rem;
        font-size: 0.8rem;
      }

      img {
        inline-size: 100%;
        block-size: 100%;
        object-fit: cover;
      }
    `]}get#t(){let t=this.name.trim().split(/\s+/).filter(Boolean);return t.length?t.length===1?t[0].slice(0,2).toUpperCase():(t[0][0]+t[t.length-1][0]).toUpperCase():"?"}#e(){this._failed=!0}updated(t){t.has("src")&&(this._failed=!1)}render(){let t=!!this.src&&!this._failed;return l`
      <span part="base" class="avatar" role=${t?h:"img"} aria-label=${t?h:this.alt||this.name||"Avatar"}>
        ${t?l`<img part="image" src=${this.src} alt=${this.alt} @error=${this.#e} />`:l`<span part="initials">${this.#t}</span>`}
      </span>
    `}};gt([n({reflect:!0})],J.prototype,"src");gt([n()],J.prototype,"alt");gt([n()],J.prototype,"name");gt([n({reflect:!0})],J.prototype,"size");gt([Z()],J.prototype,"_failed");m("mb-avatar",J);var $s=Object.defineProperty,Pe=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&$s(t,e,s),s},$t=class extends d{constructor(){super(...arguments),this.size="md",this.label="Loading"}static{this.styles=[u,p`
      :host {
        display: inline-flex;
        vertical-align: middle;
      }

      .spinner {
        border: 2px solid var(--mb-color-border);
        border-inline-end-color: var(--mb-color-accent);
        border-radius: 50%;
        animation: spin 0.7s linear infinite;
      }

      :host([size='sm']) .spinner {
        inline-size: 0.9rem;
        block-size: 0.9rem;
      }

      :host([size='md']) .spinner,
      :host(:not([size])) .spinner {
        inline-size: 1.15rem;
        block-size: 1.15rem;
      }

      @keyframes spin {
        to {
          transform: rotate(360deg);
        }
      }

      @media (prefers-reduced-motion: reduce) {
        .spinner {
          animation: none;
          border-inline-end-color: var(--mb-color-border);
          border-block-start-color: var(--mb-color-accent);
        }
      }
    `]}render(){return l`
      <span
        part="spinner"
        class="spinner"
        role="status"
        aria-label=${this.label}
      ></span>
    `}};Pe([n({reflect:!0})],$t.prototype,"size");Pe([n()],$t.prototype,"label");m("mb-spinner",$t);var Zt=class extends d{static{this.styles=[u,p`
      :host {
        display: block;
        inline-size: 100%;
      }

      .toolbar {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        justify-content: space-between;
        gap: var(--mb-space-3);
      }

      .start,
      .end {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: var(--mb-space-2);
        min-inline-size: 0;
      }

      .end {
        margin-inline-start: auto;
      }
    `]}render(){return l`
      <div part="toolbar" class="toolbar">
        <div part="start" class="start">
          <slot name="start"></slot>
          <slot></slot>
        </div>
        <div part="end" class="end">
          <slot name="end"></slot>
        </div>
      </div>
    `}};m("mb-toolbar",Zt);
/*! Bundled license information:

@lit/reactive-element/css-tag.js:
  (**
   * @license
   * Copyright 2019 Google LLC
   * SPDX-License-Identifier: BSD-3-Clause
   *)

@lit/reactive-element/reactive-element.js:
lit-html/lit-html.js:
lit-element/lit-element.js:
@lit/reactive-element/decorators/custom-element.js:
@lit/reactive-element/decorators/property.js:
@lit/reactive-element/decorators/state.js:
@lit/reactive-element/decorators/event-options.js:
@lit/reactive-element/decorators/base.js:
@lit/reactive-element/decorators/query.js:
@lit/reactive-element/decorators/query-all.js:
@lit/reactive-element/decorators/query-async.js:
@lit/reactive-element/decorators/query-assigned-nodes.js:
lit-html/directive.js:
lit-html/directives/repeat.js:
  (**
   * @license
   * Copyright 2017 Google LLC
   * SPDX-License-Identifier: BSD-3-Clause
   *)

lit-html/is-server.js:
  (**
   * @license
   * Copyright 2022 Google LLC
   * SPDX-License-Identifier: BSD-3-Clause
   *)

@lit/reactive-element/decorators/query-assigned-elements.js:
  (**
   * @license
   * Copyright 2021 Google LLC
   * SPDX-License-Identifier: BSD-3-Clause
   *)

lit-html/directive-helpers.js:
  (**
   * @license
   * Copyright 2020 Google LLC
   * SPDX-License-Identifier: BSD-3-Clause
   *)
*/
