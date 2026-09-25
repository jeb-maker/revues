var Et=globalThis,St=Et.ShadowRoot&&(Et.ShadyCSS===void 0||Et.ShadyCSS.nativeShadow)&&"adoptedStyleSheets"in Document.prototype&&"replace"in CSSStyleSheet.prototype,Jt=Symbol(),ie=new WeakMap,ht=class{constructor(t,e,i){if(this._$cssResult$=!0,i!==Jt)throw Error("CSSResult is not constructable. Use `unsafeCSS` or `css` instead.");this.cssText=t,this.t=e}get styleSheet(){let t=this.o,e=this.t;if(St&&t===void 0){let i=e!==void 0&&e.length===1;i&&(t=ie.get(e)),t===void 0&&((this.o=t=new CSSStyleSheet).replaceSync(this.cssText),i&&ie.set(e,t))}return t}toString(){return this.cssText}},oe=r=>new ht(typeof r=="string"?r:r+"",void 0,Jt),m=(r,...t)=>{let e=r.length===1?r[0]:t.reduce((i,s,o)=>i+(a=>{if(a._$cssResult$===!0)return a.cssText;if(typeof a=="number")return a;throw Error("Value passed to 'css' function must be a 'css' function result: "+a+". Use 'unsafeCSS' to pass non-literal values, but take care to ensure page security.")})(s)+r[o+1],r[0]);return new ht(e,r,Jt)},ae=(r,t)=>{if(St)r.adoptedStyleSheets=t.map(e=>e instanceof CSSStyleSheet?e:e.styleSheet);else for(let e of t){let i=document.createElement("style"),s=Et.litNonce;s!==void 0&&i.setAttribute("nonce",s),i.textContent=e.cssText,r.appendChild(i)}},Wt=St?r=>r:r=>r instanceof CSSStyleSheet?(t=>{let e="";for(let i of t.cssRules)e+=i.cssText;return oe(e)})(r):r;var{is:qe,defineProperty:Ne,getOwnPropertyDescriptor:Re,getOwnPropertyNames:Ue,getOwnPropertySymbols:Te,getPrototypeOf:Be}=Object,zt=globalThis,ne=zt.trustedTypes,je=ne?ne.emptyScript:"",Ve=zt.reactiveElementPolyfillSupport,pt=(r,t)=>r,mt={toAttribute(r,t){switch(t){case Boolean:r=r?je:null;break;case Object:case Array:r=r==null?r:JSON.stringify(r)}return r},fromAttribute(r,t){let e=r;switch(t){case Boolean:e=r!==null;break;case Number:e=r===null?null:Number(r);break;case Object:case Array:try{e=JSON.parse(r)}catch{e=null}}return e}},Ct=(r,t)=>!qe(r,t),le={attribute:!0,type:String,converter:mt,reflect:!1,useDefault:!1,hasChanged:Ct};Symbol.metadata??=Symbol("metadata"),zt.litPropertyMetadata??=new WeakMap;var B=class extends HTMLElement{static addInitializer(t){this._$Ei(),(this.l??=[]).push(t)}static get observedAttributes(){return this.finalize(),this._$Eh&&[...this._$Eh.keys()]}static createProperty(t,e=le){if(e.state&&(e.attribute=!1),this._$Ei(),this.prototype.hasOwnProperty(t)&&((e=Object.create(e)).wrapped=!0),this.elementProperties.set(t,e),!e.noAccessor){let i=Symbol(),s=this.getPropertyDescriptor(t,i,e);s!==void 0&&Ne(this.prototype,t,s)}}static getPropertyDescriptor(t,e,i){let{get:s,set:o}=Re(this.prototype,t)??{get(){return this[e]},set(a){this[e]=a}};return{get:s,set(a){let h=s?.call(this);o?.call(this,a),this.requestUpdate(t,h,i)},configurable:!0,enumerable:!0}}static getPropertyOptions(t){return this.elementProperties.get(t)??le}static _$Ei(){if(this.hasOwnProperty(pt("elementProperties")))return;let t=Be(this);t.finalize(),t.l!==void 0&&(this.l=[...t.l]),this.elementProperties=new Map(t.elementProperties)}static finalize(){if(this.hasOwnProperty(pt("finalized")))return;if(this.finalized=!0,this._$Ei(),this.hasOwnProperty(pt("properties"))){let e=this.properties,i=[...Ue(e),...Te(e)];for(let s of i)this.createProperty(s,e[s])}let t=this[Symbol.metadata];if(t!==null){let e=litPropertyMetadata.get(t);if(e!==void 0)for(let[i,s]of e)this.elementProperties.set(i,s)}this._$Eh=new Map;for(let[e,i]of this.elementProperties){let s=this._$Eu(e,i);s!==void 0&&this._$Eh.set(s,e)}this.elementStyles=this.finalizeStyles(this.styles)}static finalizeStyles(t){let e=[];if(Array.isArray(t)){let i=new Set(t.flat(1/0).reverse());for(let s of i)e.unshift(Wt(s))}else t!==void 0&&e.push(Wt(t));return e}static _$Eu(t,e){let i=e.attribute;return i===!1?void 0:typeof i=="string"?i:typeof t=="string"?t.toLowerCase():void 0}constructor(){super(),this._$Ep=void 0,this.isUpdatePending=!1,this.hasUpdated=!1,this._$Em=null,this._$Ev()}_$Ev(){this._$ES=new Promise(t=>this.enableUpdating=t),this._$AL=new Map,this._$E_(),this.requestUpdate(),this.constructor.l?.forEach(t=>t(this))}addController(t){(this._$EO??=new Set).add(t),this.renderRoot!==void 0&&this.isConnected&&t.hostConnected?.()}removeController(t){this._$EO?.delete(t)}_$E_(){let t=new Map,e=this.constructor.elementProperties;for(let i of e.keys())this.hasOwnProperty(i)&&(t.set(i,this[i]),delete this[i]);t.size>0&&(this._$Ep=t)}createRenderRoot(){let t=this.shadowRoot??this.attachShadow(this.constructor.shadowRootOptions);return ae(t,this.constructor.elementStyles),t}connectedCallback(){this.renderRoot??=this.createRenderRoot(),this.enableUpdating(!0),this._$EO?.forEach(t=>t.hostConnected?.())}enableUpdating(t){}disconnectedCallback(){this._$EO?.forEach(t=>t.hostDisconnected?.())}attributeChangedCallback(t,e,i){this._$AK(t,i)}_$ET(t,e){let i=this.constructor.elementProperties.get(t),s=this.constructor._$Eu(t,i);if(s!==void 0&&i.reflect===!0){let o=(i.converter?.toAttribute!==void 0?i.converter:mt).toAttribute(e,i.type);this._$Em=t,o==null?this.removeAttribute(s):this.setAttribute(s,o),this._$Em=null}}_$AK(t,e){let i=this.constructor,s=i._$Eh.get(t);if(s!==void 0&&this._$Em!==s){let o=i.getPropertyOptions(s),a=typeof o.converter=="function"?{fromAttribute:o.converter}:o.converter?.fromAttribute!==void 0?o.converter:mt;this._$Em=s;let h=a.fromAttribute(e,o.type);this[s]=h??this._$Ej?.get(s)??h,this._$Em=null}}requestUpdate(t,e,i,s=!1,o){if(t!==void 0){let a=this.constructor;if(s===!1&&(o=this[t]),i??=a.getPropertyOptions(t),!((i.hasChanged??Ct)(o,e)||i.useDefault&&i.reflect&&o===this._$Ej?.get(t)&&!this.hasAttribute(a._$Eu(t,i))))return;this.C(t,e,i)}this.isUpdatePending===!1&&(this._$ES=this._$EP())}C(t,e,{useDefault:i,reflect:s,wrapped:o},a){i&&!(this._$Ej??=new Map).has(t)&&(this._$Ej.set(t,a??e??this[t]),o!==!0||a!==void 0)||(this._$AL.has(t)||(this.hasUpdated||i||(e=void 0),this._$AL.set(t,e)),s===!0&&this._$Em!==t&&(this._$Eq??=new Set).add(t))}async _$EP(){this.isUpdatePending=!0;try{await this._$ES}catch(e){Promise.reject(e)}let t=this.scheduleUpdate();return t!=null&&await t,!this.isUpdatePending}scheduleUpdate(){return this.performUpdate()}performUpdate(){if(!this.isUpdatePending)return;if(!this.hasUpdated){if(this.renderRoot??=this.createRenderRoot(),this._$Ep){for(let[s,o]of this._$Ep)this[s]=o;this._$Ep=void 0}let i=this.constructor.elementProperties;if(i.size>0)for(let[s,o]of i){let{wrapped:a}=o,h=this[s];a!==!0||this._$AL.has(s)||h===void 0||this.C(s,void 0,o,h)}}let t=!1,e=this._$AL;try{t=this.shouldUpdate(e),t?(this.willUpdate(e),this._$EO?.forEach(i=>i.hostUpdate?.()),this.update(e)):this._$EM()}catch(i){throw t=!1,this._$EM(),i}t&&this._$AE(e)}willUpdate(t){}_$AE(t){this._$EO?.forEach(e=>e.hostUpdated?.()),this.hasUpdated||(this.hasUpdated=!0,this.firstUpdated(t)),this.updated(t)}_$EM(){this._$AL=new Map,this.isUpdatePending=!1}get updateComplete(){return this.getUpdateComplete()}getUpdateComplete(){return this._$ES}shouldUpdate(t){return!0}update(t){this._$Eq&&=this._$Eq.forEach(e=>this._$ET(e,this[e])),this._$EM()}updated(t){}firstUpdated(t){}};B.elementStyles=[],B.shadowRootOptions={mode:"open"},B[pt("elementProperties")]=new Map,B[pt("finalized")]=new Map,Ve?.({ReactiveElement:B}),(zt.reactiveElementVersions??=[]).push("2.1.2");var Gt=globalThis,ce=r=>r,Lt=Gt.trustedTypes,de=Lt?Lt.createPolicy("lit-html",{createHTML:r=>r}):void 0,Qt="$lit$",j=`lit$${Math.random().toFixed(9).slice(2)}$`,Xt="?"+j,He=`<${Xt}>`,Z=document,ut=()=>Z.createComment(""),ft=r=>r===null||typeof r!="object"&&typeof r!="function",Zt=Array.isArray,fe=r=>Zt(r)||typeof r?.[Symbol.iterator]=="function",Yt=`[ 	
\f\r]`,bt=/<(?:(!--|\/[^a-zA-Z])|(\/?[a-zA-Z][^>\s]*)|(\/?$))/g,he=/-->/g,pe=/>/g,Q=RegExp(`>|${Yt}(?:([^\\s"'>=/]+)(${Yt}*=${Yt}*(?:[^ 	
\f\r"'\`<>=]|("|')|))|$)`,"g"),me=/'/g,be=/"/g,ve=/^(?:script|style|textarea|title)$/i,te=r=>(t,...e)=>({_$litType$:r,strings:t,values:e}),l=te(1),Os=te(2),Ms=te(3),V=Symbol.for("lit-noChange"),c=Symbol.for("lit-nothing"),ue=new WeakMap,X=Z.createTreeWalker(Z,129);function ge(r,t){if(!Zt(r)||!r.hasOwnProperty("raw"))throw Error("invalid template strings array");return de!==void 0?de.createHTML(t):t}var ye=(r,t)=>{let e=r.length-1,i=[],s,o=t===2?"<svg>":t===3?"<math>":"",a=bt;for(let h=0;h<e;h++){let d=r[h],v,A,f=-1,$=0;for(;$<d.length&&(a.lastIndex=$,A=a.exec(d),A!==null);)$=a.lastIndex,a===bt?A[1]==="!--"?a=he:A[1]!==void 0?a=pe:A[2]!==void 0?(ve.test(A[2])&&(s=RegExp("</"+A[2],"g")),a=Q):A[3]!==void 0&&(a=Q):a===Q?A[0]===">"?(a=s??bt,f=-1):A[1]===void 0?f=-2:(f=a.lastIndex-A[2].length,v=A[1],a=A[3]===void 0?Q:A[3]==='"'?be:me):a===be||a===me?a=Q:a===he||a===pe?a=bt:(a=Q,s=void 0);let g=a===Q&&r[h+1].startsWith("/>")?" ":"";o+=a===bt?d+He:f>=0?(i.push(v),d.slice(0,f)+Qt+d.slice(f)+j+g):d+j+(f===-2?h:g)}return[ge(r,o+(r[e]||"<?>")+(t===2?"</svg>":t===3?"</math>":"")),i]},vt=class r{constructor({strings:t,_$litType$:e},i){let s;this.parts=[];let o=0,a=0,h=t.length-1,d=this.parts,[v,A]=ye(t,e);if(this.el=r.createElement(v,i),X.currentNode=this.el.content,e===2||e===3){let f=this.el.content.firstChild;f.replaceWith(...f.childNodes)}for(;(s=X.nextNode())!==null&&d.length<h;){if(s.nodeType===1){if(s.hasAttributes())for(let f of s.getAttributeNames())if(f.endsWith(Qt)){let $=A[a++],g=s.getAttribute(f).split(j),w=/([.?@])?(.*)/.exec($);d.push({type:1,index:o,name:w[2],strings:g,ctor:w[1]==="."?Dt:w[1]==="?"?Ot:w[1]==="@"?Mt:et}),s.removeAttribute(f)}else f.startsWith(j)&&(d.push({type:6,index:o}),s.removeAttribute(f));if(ve.test(s.tagName)){let f=s.textContent.split(j),$=f.length-1;if($>0){s.textContent=Lt?Lt.emptyScript:"";for(let g=0;g<$;g++)s.append(f[g],ut()),X.nextNode(),d.push({type:2,index:++o});s.append(f[$],ut())}}}else if(s.nodeType===8)if(s.data===Xt)d.push({type:2,index:o});else{let f=-1;for(;(f=s.data.indexOf(j,f+1))!==-1;)d.push({type:7,index:o}),f+=j.length-1}o++}}static createElement(t,e){let i=Z.createElement("template");return i.innerHTML=t,i}};function tt(r,t,e=r,i){if(t===V)return t;let s=i!==void 0?e._$Co?.[i]:e._$Cl,o=ft(t)?void 0:t._$litDirective$;return s?.constructor!==o&&(s?._$AO?.(!1),o===void 0?s=void 0:(s=new o(r),s._$AT(r,e,i)),i!==void 0?(e._$Co??=[])[i]=s:e._$Cl=s),s!==void 0&&(t=tt(r,s._$AS(r,t.values),s,i)),t}var Pt=class{constructor(t,e){this._$AV=[],this._$AN=void 0,this._$AD=t,this._$AM=e}get parentNode(){return this._$AM.parentNode}get _$AU(){return this._$AM._$AU}u(t){let{el:{content:e},parts:i}=this._$AD,s=(t?.creationScope??Z).importNode(e,!0);X.currentNode=s;let o=X.nextNode(),a=0,h=0,d=i[0];for(;d!==void 0;){if(a===d.index){let v;d.type===2?v=new ot(o,o.nextSibling,this,t):d.type===1?v=new d.ctor(o,d.name,d.strings,this,t):d.type===6&&(v=new qt(o,this,t)),this._$AV.push(v),d=i[++h]}a!==d?.index&&(o=X.nextNode(),a++)}return X.currentNode=Z,s}p(t){let e=0;for(let i of this._$AV)i!==void 0&&(i.strings!==void 0?(i._$AI(t,i,e),e+=i.strings.length-2):i._$AI(t[e])),e++}},ot=class r{get _$AU(){return this._$AM?._$AU??this._$Cv}constructor(t,e,i,s){this.type=2,this._$AH=c,this._$AN=void 0,this._$AA=t,this._$AB=e,this._$AM=i,this.options=s,this._$Cv=s?.isConnected??!0}get parentNode(){let t=this._$AA.parentNode,e=this._$AM;return e!==void 0&&t?.nodeType===11&&(t=e.parentNode),t}get startNode(){return this._$AA}get endNode(){return this._$AB}_$AI(t,e=this){t=tt(this,t,e),ft(t)?t===c||t==null||t===""?(this._$AH!==c&&this._$AR(),this._$AH=c):t!==this._$AH&&t!==V&&this._(t):t._$litType$!==void 0?this.$(t):t.nodeType!==void 0?this.T(t):fe(t)?this.k(t):this._(t)}O(t){return this._$AA.parentNode.insertBefore(t,this._$AB)}T(t){this._$AH!==t&&(this._$AR(),this._$AH=this.O(t))}_(t){this._$AH!==c&&ft(this._$AH)?this._$AA.nextSibling.data=t:this.T(Z.createTextNode(t)),this._$AH=t}$(t){let{values:e,_$litType$:i}=t,s=typeof i=="number"?this._$AC(t):(i.el===void 0&&(i.el=vt.createElement(ge(i.h,i.h[0]),this.options)),i);if(this._$AH?._$AD===s)this._$AH.p(e);else{let o=new Pt(s,this),a=o.u(this.options);o.p(e),this.T(a),this._$AH=o}}_$AC(t){let e=ue.get(t.strings);return e===void 0&&ue.set(t.strings,e=new vt(t)),e}k(t){Zt(this._$AH)||(this._$AH=[],this._$AR());let e=this._$AH,i,s=0;for(let o of t)s===e.length?e.push(i=new r(this.O(ut()),this.O(ut()),this,this.options)):i=e[s],i._$AI(o),s++;s<e.length&&(this._$AR(i&&i._$AB.nextSibling,s),e.length=s)}_$AR(t=this._$AA.nextSibling,e){for(this._$AP?.(!1,!0,e);t!==this._$AB;){let i=ce(t).nextSibling;ce(t).remove(),t=i}}setConnected(t){this._$AM===void 0&&(this._$Cv=t,this._$AP?.(t))}},et=class{get tagName(){return this.element.tagName}get _$AU(){return this._$AM._$AU}constructor(t,e,i,s,o){this.type=1,this._$AH=c,this._$AN=void 0,this.element=t,this.name=e,this._$AM=s,this.options=o,i.length>2||i[0]!==""||i[1]!==""?(this._$AH=Array(i.length-1).fill(new String),this.strings=i):this._$AH=c}_$AI(t,e=this,i,s){let o=this.strings,a=!1;if(o===void 0)t=tt(this,t,e,0),a=!ft(t)||t!==this._$AH&&t!==V,a&&(this._$AH=t);else{let h=t,d,v;for(t=o[0],d=0;d<o.length-1;d++)v=tt(this,h[i+d],e,d),v===V&&(v=this._$AH[d]),a||=!ft(v)||v!==this._$AH[d],v===c?t=c:t!==c&&(t+=(v??"")+o[d+1]),this._$AH[d]=v}a&&!s&&this.j(t)}j(t){t===c?this.element.removeAttribute(this.name):this.element.setAttribute(this.name,t??"")}},Dt=class extends et{constructor(){super(...arguments),this.type=3}j(t){this.element[this.name]=t===c?void 0:t}},Ot=class extends et{constructor(){super(...arguments),this.type=4}j(t){this.element.toggleAttribute(this.name,!!t&&t!==c)}},Mt=class extends et{constructor(t,e,i,s,o){super(t,e,i,s,o),this.type=5}_$AI(t,e=this){if((t=tt(this,t,e,0)??c)===V)return;let i=this._$AH,s=t===c&&i!==c||t.capture!==i.capture||t.once!==i.once||t.passive!==i.passive,o=t!==c&&(i===c||s);s&&this.element.removeEventListener(this.name,this,i),o&&this.element.addEventListener(this.name,this,t),this._$AH=t}handleEvent(t){typeof this._$AH=="function"?this._$AH.call(this.options?.host??this.element,t):this._$AH.handleEvent(t)}},qt=class{constructor(t,e,i){this.element=t,this.type=6,this._$AN=void 0,this._$AM=e,this.options=i}get _$AU(){return this._$AM._$AU}_$AI(t){tt(this,t)}},$e={M:Qt,P:j,A:Xt,C:1,L:ye,R:Pt,D:fe,V:tt,I:ot,H:et,N:Ot,U:Mt,B:Dt,F:qt},Ie=Gt.litHtmlPolyfillSupport;Ie?.(vt,ot),(Gt.litHtmlVersions??=[]).push("3.3.3");var xe=(r,t,e)=>{let i=e?.renderBefore??t,s=i._$litPart$;if(s===void 0){let o=e?.renderBefore??null;i._$litPart$=s=new ot(t.insertBefore(ut(),o),o,void 0,e??{})}return s._$AI(r),s};var ee=globalThis,p=class extends B{constructor(){super(...arguments),this.renderOptions={host:this},this._$Do=void 0}createRenderRoot(){let t=super.createRenderRoot();return this.renderOptions.renderBefore??=t.firstChild,t}update(t){let e=this.render();this.hasUpdated||(this.renderOptions.isConnected=this.isConnected),super.update(t),this._$Do=xe(e,this.renderRoot,this.renderOptions)}connectedCallback(){super.connectedCallback(),this._$Do?.setConnected(!0)}disconnectedCallback(){super.disconnectedCallback(),this._$Do?.setConnected(!1)}render(){return V}};p._$litElement$=!0,p.finalized=!0,ee.litElementHydrateSupport?.({LitElement:p});var Ke=ee.litElementPolyfillSupport;Ke?.({LitElement:p});(ee.litElementVersions??=[]).push("4.2.2");var Fe={attribute:!0,type:String,converter:mt,reflect:!1,hasChanged:Ct},Je=(r=Fe,t,e)=>{let{kind:i,metadata:s}=e,o=globalThis.litPropertyMetadata.get(s);if(o===void 0&&globalThis.litPropertyMetadata.set(s,o=new Map),i==="setter"&&((r=Object.create(r)).wrapped=!0),o.set(e.name,r),i==="accessor"){let{name:a}=e;return{set(h){let d=t.get.call(this);t.set.call(this,h),this.requestUpdate(a,d,r,!0,h)},init(h){return h!==void 0&&this.C(a,void 0,r,h),h}}}if(i==="setter"){let{name:a}=e;return function(h){let d=this[a];t.call(this,h),this.requestUpdate(a,d,r,!0,h)}}throw Error("Unsupported decorator location: "+i)};function n(r){return(t,e)=>typeof e=="object"?Je(r,t,e):((i,s,o)=>{let a=s.hasOwnProperty(o);return s.constructor.createProperty(o,i),a?Object.getOwnPropertyDescriptor(s,o):void 0})(r,t,e)}function H(r){return n({...r,state:!0,attribute:!1})}function _(r,t,e){r.setFormValue(t,t)}function R(r,t,e="",i){r.setValidity(t,e,i)}function U(r){r.setValidity({})}function K(r,t,e="Please fill out this field."){return r?{flags:{customError:!0},message:r}:t?{flags:{valueMissing:!0},message:e}:{flags:{},message:""}}function b(r,t){customElements.get(r)||customElements.define(r,t)}var u=m`
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
`,at=m`
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
`;function nt(r,t,e){return r?{labelText:r,hideVisually:t,controlAriaLabel:""}:{labelText:"",hideVisually:!1,controlAriaLabel:e}}var We=Object.defineProperty,M=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&We(t,e,s),s},C=class extends p{constructor(){super(...arguments),this.variant="primary",this.size="md",this.type="button",this.disabled=!1,this.loading=!1,this.name="",this.value="",this.href="",this.target="",this.rel="",this.iconOnly=!1,this.#t=this.attachInternals(),this.#e=!1}static{this.formAssociated=!0}static{this.styles=[u,m`
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
    `]}#t;#e;get#s(){return this.disabled||this.loading||this.#e}get#r(){return!!this.href}get#o(){return this.getAttribute("aria-label")??""}formDisabledCallback(t){this.#e=t,this.requestUpdate()}#i(t){if(this.#s){t.preventDefault(),t.stopImmediatePropagation();return}if(this.#r)return;let e=this.#t.form;e&&(this.type==="submit"?(this.name&&_(this.#t,this.value),e.requestSubmit(),queueMicrotask(()=>_(this.#t,null))):this.type==="reset"&&e.reset())}render(){let t=l`
      ${this.loading?l`<span class="spinner" aria-hidden="true"></span>`:c}
      <slot></slot>
    `,e=this.#o||c;return this.#r?l`
        <a
          part="base"
          class="base"
          href=${this.#s?c:this.href}
          target=${this.target||c}
          rel=${this.rel||(this.target==="_blank"?"noopener noreferrer":c)}
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
    `}};M([n({reflect:!0})],C.prototype,"variant");M([n({reflect:!0})],C.prototype,"size");M([n({reflect:!0})],C.prototype,"type");M([n({type:Boolean,reflect:!0})],C.prototype,"disabled");M([n({type:Boolean,reflect:!0})],C.prototype,"loading");M([n({reflect:!0})],C.prototype,"name");M([n()],C.prototype,"value");M([n({reflect:!0})],C.prototype,"href");M([n({reflect:!0})],C.prototype,"target");M([n({reflect:!0})],C.prototype,"rel");M([n({type:Boolean,reflect:!0,attribute:"icon-only"})],C.prototype,"iconOnly");b("mb-button",C);var Ye=Object.defineProperty,Ge=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Ye(t,e,s),s},Rt=class extends p{constructor(){super(...arguments),this.variant="neutral"}static{this.styles=[u,m`
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
    `]}render(){return l`<span part="base"><slot></slot></span>`}};Ge([n({reflect:!0})],Rt.prototype,"variant");b("mb-badge",Rt);var Qe=Object.defineProperty,Xe=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Qe(t,e,s),s},Ut=class extends p{constructor(){super(...arguments),this.variant="info"}static{this.styles=[u,m`
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
    `}};Xe([n({reflect:!0})],Ut.prototype,"variant");b("mb-alert",Ut);var Ze=Object.defineProperty,Ae=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Ze(t,e,s),s},gt=class extends p{constructor(){super(...arguments),this._hasHeader=!1,this._hasFooter=!1}static{this.styles=[u,m`
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
    `}};Ae([H()],gt.prototype,"_hasHeader");Ae([H()],gt.prototype,"_hasFooter");b("mb-card",gt);var ts=Object.defineProperty,k=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&ts(t,e,s),s},x=class extends p{constructor(){super(...arguments),this.label="",this.hint="",this.error="",this.value="",this.name="",this.placeholder="",this.type="text",this.disabled=!1,this.required=!1,this.invalid=!1,this.density="default",this.hideLabel=!1,this.min="",this.max="",this.step="",this.accept="",this.multiple=!1,this.#t=this.attachInternals(),this.#e=!1,this.#r="",this.#o=!1,this.#i=!1}static{this.formAssociated=!0}static{this.styles=[u,at,m`
      :host {
        display: block;
      }

      input[type='file'].control {
        padding-block: var(--mb-space-2);
      }
    `]}#t;#e;#s;#r;#o;#i;get#l(){return this.disabled||this.#e}get#a(){return this.type==="file"}get#n(){return this.getAttribute("aria-label")??""}connectedCallback(){super.connectedCallback(),this.#o||(this.#r=this.value,this.#o=!0)}firstUpdated(){this.#s=this.renderRoot.querySelector("input")??void 0,this.#c()}updated(t){(t.has("value")||t.has("required")||t.has("error")||t.has("disabled")||t.has("name")||t.has("type"))&&this.#c()}formDisabledCallback(t){this.#e=t,this.requestUpdate()}formResetCallback(){this.#i=!1,this.value=this.#r,this.error="",this.invalid=!1,this.#a&&this.#s&&(this.#s.value="")}#d(){let t=this.#s?.files;if(!this.name||!t?.length){_(this.#t,null);return}if(t.length===1){_(this.#t,t[0]);return}let e=new FormData;for(let i of t)e.append(this.name,i);_(this.#t,e)}#c(){this.#a?this.#d():_(this.#t,this.name?this.value:null);let t=this.required&&(this.#a?!this.#s?.files?.length:!this.value),{flags:e,message:i}=K(this.error,t);i?(R(this.#t,e,i,this.#s),this.invalid=!!this.error||this.#i):(U(this.#t),this.invalid=!1)}#p(t){let e=t.target;this.#i=!0,this.#a||(this.value=e.value),this.#c(),this.dispatchEvent(new CustomEvent("mb-input",{detail:{value:this.value,files:e.files},bubbles:!0,composed:!0}))}#h(t){let e=t.target;this.#i=!0,this.#a||(this.value=e.value),this.#c(),this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value,files:e.files},bubbles:!0,composed:!0}))}#m(t){if(t.key!=="Enter"||t.defaultPrevented||this.#a)return;let e=this.#t.form;e&&(t.preventDefault(),e.requestSubmit())}render(){let t=[this.hint&&!this.error?"hint":"",this.error?"error":""].filter(Boolean).join(" "),{labelText:e,hideVisually:i,controlAriaLabel:s}=nt(this.label,this.hideLabel,this.#n);return l`
      <div class="field">
        ${e?l`<label
              part="label"
              class="label${i?" visually-hidden":""}"
              for="control"
              >${e}</label
            >`:c}
        <input
          id="control"
          part="control"
          class="control"
          .type=${this.type}
          .value=${this.#a?"":this.value}
          name=${this.name||c}
          placeholder=${this.placeholder||c}
          min=${this.type==="number"&&this.min!==""?this.min:c}
          max=${this.type==="number"&&this.max!==""?this.max:c}
          step=${this.type==="number"&&this.step!==""?this.step:c}
          accept=${this.#a&&this.accept?this.accept:c}
          ?multiple=${this.#a&&this.multiple}
          ?disabled=${this.#l}
          ?required=${this.required}
          aria-invalid=${this.invalid?"true":"false"}
          aria-label=${s||c}
          aria-describedby=${t||c}
          @input=${this.#p}
          @change=${this.#h}
          @keydown=${this.#m}
        />
        ${this.hint&&!this.error?l`<p id="hint" class="hint">${this.hint}</p>`:c}
        ${this.error?l`<p id="error" class="error" role="alert">${this.error}</p>`:c}
      </div>
    `}};k([n()],x.prototype,"label");k([n()],x.prototype,"hint");k([n()],x.prototype,"error");k([n()],x.prototype,"value");k([n({reflect:!0})],x.prototype,"name");k([n()],x.prototype,"placeholder");k([n({reflect:!0})],x.prototype,"type");k([n({type:Boolean,reflect:!0})],x.prototype,"disabled");k([n({type:Boolean,reflect:!0})],x.prototype,"required");k([n({type:Boolean,reflect:!0})],x.prototype,"invalid");k([n({reflect:!0})],x.prototype,"density");k([n({type:Boolean,reflect:!0,attribute:"hide-label"})],x.prototype,"hideLabel");k([n()],x.prototype,"min");k([n()],x.prototype,"max");k([n()],x.prototype,"step");k([n()],x.prototype,"accept");k([n({type:Boolean})],x.prototype,"multiple");b("mb-input",x);var es=Object.defineProperty,D=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&es(t,e,s),s},z=class extends p{constructor(){super(...arguments),this.label="",this.hint="",this.error="",this.value="",this.name="",this.placeholder="",this.disabled=!1,this.required=!1,this.invalid=!1,this.rows=4,this.density="default",this.hideLabel=!1,this.#t=this.attachInternals(),this.#e=!1,this.#r="",this.#o=!1,this.#i=!1}static{this.formAssociated=!0}static{this.styles=[u,at,m`
      :host {
        display: block;
      }

      textarea.control {
        min-block-size: 6rem;
        resize: vertical;
      }
    `]}#t;#e;#s;#r;#o;#i;get#l(){return this.disabled||this.#e}get#a(){return this.getAttribute("aria-label")??""}connectedCallback(){super.connectedCallback(),this.#o||(this.#r=this.value,this.#o=!0)}firstUpdated(){this.#s=this.renderRoot.querySelector("textarea")??void 0,this.#n()}updated(t){(t.has("value")||t.has("required")||t.has("error")||t.has("disabled")||t.has("name"))&&this.#n()}formDisabledCallback(t){this.#e=t,this.requestUpdate()}formResetCallback(){this.#i=!1,this.value=this.#r,this.error="",this.invalid=!1}#n(){_(this.#t,this.name?this.value:null);let t=this.required&&!this.value,{flags:e,message:i}=K(this.error,t);i?(R(this.#t,e,i,this.#s),this.invalid=!!this.error||this.#i):(U(this.#t),this.invalid=!1)}#d(t){let e=t.target;this.#i=!0,this.value=e.value,this.dispatchEvent(new CustomEvent("mb-input",{detail:{value:this.value},bubbles:!0,composed:!0}))}#c(t){let e=t.target;this.#i=!0,this.value=e.value,this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value},bubbles:!0,composed:!0}))}render(){let t=[this.hint&&!this.error?"hint":"",this.error?"error":""].filter(Boolean).join(" "),{labelText:e,hideVisually:i,controlAriaLabel:s}=nt(this.label,this.hideLabel,this.#a);return l`
      <div class="field">
        ${e?l`<label
              part="label"
              class="label${i?" visually-hidden":""}"
              for="control"
              >${e}</label
            >`:c}
        <textarea
          id="control"
          part="control"
          class="control"
          .value=${this.value}
          name=${this.name||c}
          placeholder=${this.placeholder||c}
          rows=${this.rows}
          ?disabled=${this.#l}
          ?required=${this.required}
          aria-invalid=${this.invalid?"true":"false"}
          aria-label=${s||c}
          aria-describedby=${t||c}
          @input=${this.#d}
          @change=${this.#c}
        ></textarea>
        ${this.hint&&!this.error?l`<p id="hint" class="hint">${this.hint}</p>`:c}
        ${this.error?l`<p id="error" class="error" role="alert">${this.error}</p>`:c}
      </div>
    `}};D([n()],z.prototype,"label");D([n()],z.prototype,"hint");D([n()],z.prototype,"error");D([n()],z.prototype,"value");D([n({reflect:!0})],z.prototype,"name");D([n()],z.prototype,"placeholder");D([n({type:Boolean,reflect:!0})],z.prototype,"disabled");D([n({type:Boolean,reflect:!0})],z.prototype,"required");D([n({type:Boolean,reflect:!0})],z.prototype,"invalid");D([n({type:Number})],z.prototype,"rows");D([n({reflect:!0})],z.prototype,"density");D([n({type:Boolean,reflect:!0,attribute:"hide-label"})],z.prototype,"hideLabel");b("mb-textarea",z);var ss=Object.defineProperty,I=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&ss(t,e,s),s},O=class extends p{constructor(){super(...arguments),this.label="",this.error="",this.name="",this.value="on",this.checked=!1,this.indeterminate=!1,this.disabled=!1,this.required=!1,this.invalid=!1,this.#t=this.attachInternals(),this.#e=!1,this.#r=!1,this.#o=!1,this.#i=!1}static{this.formAssociated=!0}static{this.styles=[u,m`
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
    `]}#t;#e;#s;#r;#o;#i;get#l(){return this.disabled||this.#e}connectedCallback(){super.connectedCallback(),this.#o||(this.#r=this.checked,this.#o=!0)}firstUpdated(){this.#s=this.renderRoot.querySelector("input")??void 0,this.#a(),this.#n()}updated(t){t.has("indeterminate")&&this.#a(),(t.has("checked")||t.has("value")||t.has("required")||t.has("error")||t.has("disabled")||t.has("name"))&&this.#n()}formDisabledCallback(t){this.#e=t,this.requestUpdate()}formResetCallback(){this.#i=!1,this.checked=this.#r,this.indeterminate=!1,this.error="",this.invalid=!1}#a(){this.#s&&(this.#s.indeterminate=this.indeterminate)}#n(){_(this.#t,this.name&&this.checked?this.value:null);let t=this.required&&!this.checked,e=this.error||(t?"Please check this box.":"");if(e){let i=this.error?{customError:!0}:{valueMissing:!0};R(this.#t,i,e,this.#s),this.invalid=!!this.error||this.#i}else U(this.#t),this.invalid=!1}#d(t){let e=t.target;this.#i=!0,this.checked=e.checked,this.indeterminate=!1,this.dispatchEvent(new CustomEvent("mb-change",{detail:{checked:this.checked,value:this.value},bubbles:!0,composed:!0}))}render(){let t=this.error?"error":"";return l`
      <label part="label">
        <input
          part="control"
          type="checkbox"
          .checked=${this.checked}
          name=${this.name||c}
          value=${this.value}
          ?disabled=${this.#l}
          ?required=${this.required}
          aria-invalid=${this.invalid?"true":"false"}
          aria-describedby=${t||c}
          @change=${this.#d}
        />
        <span>${this.label}<slot></slot></span>
      </label>
      ${this.error?l`<p id="error" class="error" role="alert">${this.error}</p>`:c}
    `}};I([n()],O.prototype,"label");I([n()],O.prototype,"error");I([n({reflect:!0})],O.prototype,"name");I([n()],O.prototype,"value");I([n({type:Boolean,reflect:!0})],O.prototype,"checked");I([n({type:Boolean,reflect:!0})],O.prototype,"indeterminate");I([n({type:Boolean,reflect:!0})],O.prototype,"disabled");I([n({type:Boolean,reflect:!0})],O.prototype,"required");I([n({type:Boolean,reflect:!0})],O.prototype,"invalid");b("mb-checkbox",O);var we={ATTRIBUTE:1,CHILD:2,PROPERTY:3,BOOLEAN_ATTRIBUTE:4,EVENT:5,ELEMENT:6},ke=r=>(...t)=>({_$litDirective$:r,values:t}),Tt=class{constructor(t){}get _$AU(){return this._$AM._$AU}_$AT(t,e,i){this._$Ct=t,this._$AM=e,this._$Ci=i}_$AS(t,e){return this.update(t,e)}update(t,e){return this.render(...e)}};var{I:rs}=$e,_e=r=>r;var Ee=()=>document.createComment(""),lt=(r,t,e)=>{let i=r._$AA.parentNode,s=t===void 0?r._$AB:t._$AA;if(e===void 0){let o=i.insertBefore(Ee(),s),a=i.insertBefore(Ee(),s);e=new rs(o,a,r,r.options)}else{let o=e._$AB.nextSibling,a=e._$AM,h=a!==r;if(h){let d;e._$AQ?.(r),e._$AM=r,e._$AP!==void 0&&(d=r._$AU)!==a._$AU&&e._$AP(d)}if(o!==s||h){let d=e._$AA;for(;d!==o;){let v=_e(d).nextSibling;_e(i).insertBefore(d,s),d=v}}}return e},F=(r,t,e=r)=>(r._$AI(t,e),r),is={},Se=(r,t=is)=>r._$AH=t,ze=r=>r._$AH,Bt=r=>{r._$AR(),r._$AA.remove()};var Ce=(r,t,e)=>{let i=new Map;for(let s=t;s<=e;s++)i.set(r[s],s);return i},ct=ke(class extends Tt{constructor(r){if(super(r),r.type!==we.CHILD)throw Error("repeat() can only be used in text expressions")}dt(r,t,e){let i;e===void 0?e=t:t!==void 0&&(i=t);let s=[],o=[],a=0;for(let h of r)s[a]=i?i(h,a):a,o[a]=e(h,a),a++;return{values:o,keys:s}}render(r,t,e){return this.dt(r,t,e).values}update(r,[t,e,i]){let s=ze(r),{values:o,keys:a}=this.dt(t,e,i);if(!Array.isArray(s))return this.ut=a,o;let h=this.ut??=[],d=[],v,A,f=0,$=s.length-1,g=0,w=o.length-1;for(;f<=$&&g<=w;)if(s[f]===null)f++;else if(s[$]===null)$--;else if(h[f]===a[g])d[g]=F(s[f],o[g]),f++,g++;else if(h[$]===a[w])d[w]=F(s[$],o[w]),$--,w--;else if(h[f]===a[w])d[w]=F(s[f],o[w]),lt(r,d[w+1],s[f]),f++,w--;else if(h[$]===a[g])d[g]=F(s[$],o[g]),lt(r,s[f],s[$]),$--,g++;else if(v===void 0&&(v=Ce(a,g,w),A=Ce(h,f,$)),v.has(h[f]))if(v.has(h[$])){let T=A.get(a[g]),Ft=T!==void 0?s[T]:null;if(Ft===null){let re=lt(r,s[f]);F(re,o[g]),d[g]=re}else d[g]=F(Ft,o[g]),lt(r,s[f],Ft),s[T]=null;g++}else Bt(s[$]),$--;else Bt(s[f]),f++;for(;g<=w;){let T=lt(r,d[w+1]);F(T,o[g]),d[g++]=T}for(;f<=$;){let T=s[f++];T!==null&&Bt(T)}return this.ut=a,Se(r,d),V}});var os=Object.defineProperty,L=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&os(t,e,s),s};function as(r){if(!r)return[];try{let t=JSON.parse(r);return Array.isArray(t)?t.filter(e=>!!e&&typeof e=="object"&&typeof e.value=="string"&&typeof e.label=="string").map(e=>({value:e.value,label:e.label,disabled:!!e.disabled})):[]}catch{return[]}}var E=class extends p{constructor(){super(...arguments),this.label="",this.hint="",this.error="",this.value="",this.name="",this.disabled=!1,this.required=!1,this.invalid=!1,this.density="default",this.hideLabel=!1,this.placeholder="",this.options=[],this._slottedOptions=[],this.#t=this.attachInternals(),this.#e=!1,this.#r="",this.#o=!1,this.#i=!1}static{this.formAssociated=!0}static{this.styles=[u,at,m`
      :host {
        display: block;
      }

      /* Options are mirrored into the shadow <select>; keep light-DOM slots invisible. */
      slot {
        display: none;
      }
    `]}#t;#e;#s;#r;#o;#i;get#l(){return this.disabled||this.#e}get#a(){return this._slottedOptions.length?this._slottedOptions:this.options}get#n(){return this.#a.filter(t=>t.value!=="")}get#d(){let t=this.#a.find(e=>e.value==="");return t?.label?t.label:this.placeholder}get#c(){return this.getAttribute("aria-label")??""}connectedCallback(){super.connectedCallback(),this.#o||(this.#r=this.value,this.#o=!0),this.#h()}firstUpdated(){this.#s=this.renderRoot.querySelector("select")??void 0,this.#b()}updated(t){(t.has("value")||t.has("required")||t.has("error")||t.has("options")||t.has("_slottedOptions")||t.has("disabled")||t.has("name"))&&this.#b()}formDisabledCallback(t){this.#e=t,this.requestUpdate()}formResetCallback(){this.#i=!1,this.value=this.#r,this.error="",this.invalid=!1}#p(t){return t instanceof HTMLOptionElement?{value:t.value,label:t.label||t.textContent?.trim()||t.value,disabled:t.disabled}:null}#h(){let t=[...this.querySelectorAll(":scope > option")].map(e=>this.#p(e)).filter(e=>e!=null);t.length&&(this._slottedOptions=t)}#m(){let t=this.renderRoot.querySelector('slot[name="options"]'),e=this.renderRoot.querySelector("slot:not([name])"),i=[...t?.assignedElements({flatten:!0})??[],...e?.assignedElements({flatten:!0})??[]].map(a=>this.#p(a)).filter(a=>a!=null),s=JSON.stringify(this._slottedOptions),o=JSON.stringify(i);s!==o&&(this._slottedOptions=i)}#v(){this.#m()}#b(){this.#s&&this.#s.value!==this.value&&(this.#s.value=this.value),_(this.#t,this.name?this.value:null);let t=this.required&&!this.value,{flags:e,message:i}=K(this.error,t,"Please select an option.");i?(R(this.#t,e,i,this.#s),this.invalid=!!this.error||this.#i):(U(this.#t),this.invalid=!1)}#$(t){let e=t.target;this.#i=!0,this.value=e.value,this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value},bubbles:!0,composed:!0}))}render(){let t=[this.hint&&!this.error?"hint":"",this.error?"error":""].filter(Boolean).join(" "),{labelText:e,hideVisually:i,controlAriaLabel:s}=nt(this.label,this.hideLabel,this.#c);return l`
      <div class="field">
        ${e?l`<label
              part="label"
              class="label${i?" visually-hidden":""}"
              for="control"
              >${e}</label
            >`:c}
        <select
          id="control"
          part="control"
          class="control"
          name=${this.name||c}
          ?disabled=${this.#l}
          ?required=${this.required}
          aria-invalid=${this.invalid?"true":"false"}
          aria-label=${s||c}
          aria-describedby=${t||c}
          .value=${this.value}
          @change=${this.#$}
        >
          <option value="" ?disabled=${this.required}>${this.#d}</option>
          ${ct(this.#n,o=>o.value,o=>l`
              <option value=${o.value} ?disabled=${!!o.disabled}>
                ${o.label}
              </option>
            `)}
        </select>
        ${this.hint&&!this.error?l`<p id="hint" class="hint">${this.hint}</p>`:c}
        ${this.error?l`<p id="error" class="error" role="alert">${this.error}</p>`:c}
      </div>
      <slot name="options" @slotchange=${this.#v}></slot>
      <slot @slotchange=${this.#v}></slot>
    `}};L([n()],E.prototype,"label");L([n()],E.prototype,"hint");L([n()],E.prototype,"error");L([n()],E.prototype,"value");L([n({reflect:!0})],E.prototype,"name");L([n({type:Boolean,reflect:!0})],E.prototype,"disabled");L([n({type:Boolean,reflect:!0})],E.prototype,"required");L([n({type:Boolean,reflect:!0})],E.prototype,"invalid");L([n({reflect:!0})],E.prototype,"density");L([n({type:Boolean,reflect:!0,attribute:"hide-label"})],E.prototype,"hideLabel");L([n()],E.prototype,"placeholder");L([n({attribute:"options",converter:{fromAttribute:as,toAttribute(r){return r?.length?JSON.stringify(r):null}}})],E.prototype,"options");L([H()],E.prototype,"_slottedOptions");b("mb-select",E);var ns=Object.defineProperty,Le=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&ns(t,e,s),s},yt=class extends p{constructor(){super(...arguments),this.open=!1,this.heading="",this.#e=!1}static{this.styles=[u,m`
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
    `}};Le([n({type:Boolean,reflect:!0})],yt.prototype,"open");Le([n()],yt.prototype,"heading");b("mb-modal",yt);var ls=Object.defineProperty,jt=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&ls(t,e,s),s},st=class extends p{constructor(){super(...arguments),this.value=0,this.max=100,this.percent=null,this.label=""}static{this.styles=[u,m`
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
        ${this.label?l`<div part="label" class="label" id="label">${this.label}</div>`:c}
        <div
          part="track"
          class="track"
          role="progressbar"
          aria-valuemin="0"
          aria-valuenow=${this.#e}
          aria-valuemax=${this.#s}
          aria-labelledby=${this.label?"label":c}
          aria-label=${this.label?c:this.getAttribute("aria-label")||"Progress"}
        >
          <div part="bar" class="bar" style="inline-size: ${t}%"></div>
        </div>
        <slot></slot>
      </div>
    `}};jt([n({type:Number})],st.prototype,"value");jt([n({type:Number})],st.prototype,"max");jt([n({type:Number})],st.prototype,"percent");jt([n()],st.prototype,"label");b("mb-progress",st);var cs=Object.defineProperty,ds=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&cs(t,e,s),s},Vt=class extends p{constructor(){super(...arguments),this.label="Filters"}static{this.styles=[u,m`
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
    `}};ds([n()],Vt.prototype,"label");b("mb-segmented-control",Vt);var hs=Object.defineProperty,ps=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&hs(t,e,s),s},Ht=class extends p{constructor(){super(...arguments),this.heading=""}static{this.styles=[u,m`
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
    `}};ps([n()],Ht.prototype,"heading");b("mb-empty-state",Ht);var ms=Object.defineProperty,J=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&ms(t,e,s),s},q=class extends p{constructor(){super(...arguments),this.prevUrl="",this.nextUrl="",this.prevDisabled=!1,this.nextDisabled=!1,this.status="",this.prevLabel="Previous",this.nextLabel="Next",this.label="Pagination"}static{this.styles=[u,m`
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
    `}};J([n({attribute:"prev-url"})],q.prototype,"prevUrl");J([n({attribute:"next-url"})],q.prototype,"nextUrl");J([n({type:Boolean,attribute:"prev-disabled"})],q.prototype,"prevDisabled");J([n({type:Boolean,attribute:"next-disabled"})],q.prototype,"nextDisabled");J([n()],q.prototype,"status");J([n({attribute:"prev-label"})],q.prototype,"prevLabel");J([n({attribute:"next-label"})],q.prototype,"nextLabel");J([n()],q.prototype,"label");b("mb-pagination",q);var bs=Object.defineProperty,It=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&bs(t,e,s),s},rt=class extends p{constructor(){super(...arguments),this.open=!1,this.variant="info",this.autoDismiss=4e3,this.message="",this.#t=0,this.#e=t=>{let e=t.detail;e&&(e.variant&&(this.variant=e.variant),e.message!=null&&(this.message=e.message),e.autoDismiss!=null&&(this.autoDismiss=e.autoDismiss),this.show())}}static{this.styles=[u,m`
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
    `}};It([n({type:Boolean,reflect:!0})],rt.prototype,"open");It([n({reflect:!0})],rt.prototype,"variant");It([n({type:Number,attribute:"auto-dismiss"})],rt.prototype,"autoDismiss");It([n()],rt.prototype,"message");b("mb-toast",rt);var us=Object.defineProperty,$t=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&us(t,e,s),s},W=class extends p{constructor(){super(...arguments),this.value="",this.label="",this.disabled=!1,this.checked=!1,this.name=""}static{this.styles=[u,m`
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
          name=${this.name||c}
          .value=${this.value}
          .checked=${this.checked}
          ?disabled=${this.disabled}
          @change=${this.#e}
        />
        <span>${this.label}<slot></slot></span>
      </label>
    `}};$t([n()],W.prototype,"value");$t([n()],W.prototype,"label");$t([n({type:Boolean,reflect:!0})],W.prototype,"disabled");$t([n({type:Boolean,reflect:!0})],W.prototype,"checked");$t([n()],W.prototype,"name");b("mb-radio",W);var fs=Object.defineProperty,Y=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&fs(t,e,s),s};function vs(r){if(!r)return[];try{let t=JSON.parse(r);return Array.isArray(t)?t.filter(e=>!!e&&typeof e=="object"&&typeof e.value=="string"&&typeof e.label=="string").map(e=>({value:e.value,label:e.label,disabled:!!e.disabled})):[]}catch{return[]}}var N=class extends p{constructor(){super(...arguments),this.label="",this.error="",this.value="",this.name="",this.disabled=!1,this.required=!1,this.invalid=!1,this.options=[],this.#t=this.attachInternals(),this.#e=!1,this.#s="",this.#r=!1,this.#o=!1,this.#i=new WeakMap,this.#p=t=>{let e=t.detail?.value;e!=null&&(this.#o=!0,this.value=e,this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value},bubbles:!0,composed:!0})))},this.#h=t=>{if(!["ArrowDown","ArrowUp","ArrowRight","ArrowLeft"].includes(t.key))return;let e=this.#n().filter(a=>!a.disabled);if(!e.length)return;t.preventDefault();let i=e.findIndex(a=>a.value===this.value),s=t.key==="ArrowDown"||t.key==="ArrowRight"?1:-1,o=e[(i+s+e.length)%e.length];this.#o=!0,this.value=o.value,o.focus(),this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value},bubbles:!0,composed:!0}))}}static{this.formAssociated=!0}static{this.styles=[u,m`
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
    `]}#t;#e;#s;#r;#o;#i;get#l(){return this.disabled||this.#e}connectedCallback(){super.connectedCallback(),this.#r||(this.#s=this.value,this.#r=!0),this.addEventListener("mb-radio-select",this.#p),this.addEventListener("keydown",this.#h)}disconnectedCallback(){super.disconnectedCallback(),this.removeEventListener("mb-radio-select",this.#p),this.removeEventListener("keydown",this.#h)}firstUpdated(){this.#d(),this.#c()}updated(t){(t.has("value")||t.has("name")||t.has("disabled")||t.has("options"))&&this.#d(),(t.has("value")||t.has("required")||t.has("error")||t.has("name")||t.has("disabled"))&&this.#c()}formDisabledCallback(t){this.#e=t,this.requestUpdate(),this.#d()}formResetCallback(){this.#o=!1,this.value=this.#s,this.error="",this.invalid=!1}#a(){return this.renderRoot.querySelector("slot")?.assignedElements({flatten:!0}).filter(t=>t.localName==="mb-radio")??[]}#n(){let t=[...this.renderRoot.querySelectorAll(".options > mb-radio")];return[...this.#a(),...t]}#d(){let t=this.#n();for(let e of t)e.name=this.name||"mb-radio-group",e.checked=e.value===this.value;for(let e of this.#a())this.#i.has(e)||this.#i.set(e,e.disabled),e.disabled=this.#l||!!this.#i.get(e)}#c(){_(this.#t,this.name?this.value:null);let t=this.required&&!this.value,{flags:e,message:i}=K(this.error,t,"Please select an option.");i?(R(this.#t,e,i),this.invalid=!!this.error||this.#o):(U(this.#t),this.invalid=!1)}#p;#h;#m(){this.#d()}render(){return l`
      <fieldset part="fieldset" ?disabled=${this.#l}>
        ${this.label?l`<legend part="legend">${this.label}</legend>`:c}
        <div class="options" part="options" role="radiogroup" aria-invalid=${this.invalid?"true":"false"}>
          <slot @slotchange=${this.#m}></slot>
          ${this.options.map(t=>l`
              <mb-radio
                .value=${t.value}
                .label=${t.label}
                ?disabled=${!!t.disabled||this.#l}
                ?checked=${t.value===this.value}
                .name=${this.name||"mb-radio-group"}
              ></mb-radio>
            `)}
        </div>
        ${this.error?l`<p class="error" role="alert">${this.error}</p>`:c}
      </fieldset>
    `}};Y([n()],N.prototype,"label");Y([n()],N.prototype,"error");Y([n()],N.prototype,"value");Y([n({reflect:!0})],N.prototype,"name");Y([n({type:Boolean,reflect:!0})],N.prototype,"disabled");Y([n({type:Boolean,reflect:!0})],N.prototype,"required");Y([n({type:Boolean,reflect:!0})],N.prototype,"invalid");Y([n({attribute:"options",converter:{fromAttribute:vs,toAttribute(r){return r?.length?JSON.stringify(r):null}}})],N.prototype,"options");b("mb-radio-group",N);var gs=Object.defineProperty,Pe=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&gs(t,e,s),s},xt=class extends p{constructor(){super(...arguments),this.href="",this.size="md"}static{this.styles=[u,m`
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
      `:l`<span part="base" class="tag"><slot></slot></span>`}};Pe([n({reflect:!0})],xt.prototype,"href");Pe([n({reflect:!0})],xt.prototype,"size");b("mb-tag",xt);var ys=Object.defineProperty,De=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&ys(t,e,s),s};function $s(r){if(!r)return[];try{let t=JSON.parse(r);return Array.isArray(t)?t.filter(e=>!!e&&typeof e=="object"&&typeof e.label=="string").map(e=>({label:e.label,href:e.href,current:!!e.current})):[]}catch{return[]}}var At=class extends p{constructor(){super(...arguments),this.label="Breadcrumb",this.items=[]}static{this.styles=[u,m`
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
          ${this.items.length?ct(this.items,t=>`${t.href??""}:${t.label}`,t=>l`
                  <li part="item">
                    ${t.current||!t.href?l`<span aria-current=${t.current?"page":c}
                          >${t.label}</span
                        >`:l`<a href=${t.href}>${t.label}</a>`}
                  </li>
                `):l`<slot></slot>`}
        </ol>
      </nav>
    `}};De([n()],At.prototype,"label");De([n({attribute:"items",converter:{fromAttribute:$s,toAttribute(r){return r?.length?JSON.stringify(r):null}}})],At.prototype,"items");b("mb-breadcrumbs",At);var xs=Object.defineProperty,Oe=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&xs(t,e,s),s},wt=class extends p{constructor(){super(...arguments),this.label="Primary",this.open=!1}static{this.styles=[u,m`
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
    `}};Oe([n()],wt.prototype,"label");Oe([n({type:Boolean,reflect:!0})],wt.prototype,"open");b("mb-nav",wt);var As=Object.defineProperty,Kt=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&As(t,e,s),s},it=class extends p{constructor(){super(...arguments),this.expanded=!1,this.for="",this.labelOpen="Menu",this.labelClose="Close menu"}static{this.styles=[u,m`
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
        aria-controls=${this.for||c}
        aria-label=${this.expanded?this.labelClose:this.labelOpen}
        @click=${this.#t}
      >
        <slot>${this.expanded?"\u2715":"\u2630"}</slot>
      </button>
    `}};Kt([n({type:Boolean,reflect:!0})],it.prototype,"expanded");Kt([n({attribute:"for"})],it.prototype,"for");Kt([n({attribute:"label-open"})],it.prototype,"labelOpen");Kt([n({attribute:"label-close"})],it.prototype,"labelClose");b("mb-nav-toggle",it);var ws=Object.defineProperty,kt=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&ws(t,e,s),s},G=class extends p{constructor(){super(...arguments),this.src="",this.alt="",this.name="",this.size="md",this._failed=!1}static{this.styles=[u,m`
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
      <span part="base" class="avatar" role=${t?c:"img"} aria-label=${t?c:this.alt||this.name||"Avatar"}>
        ${t?l`<img part="image" src=${this.src} alt=${this.alt} @error=${this.#e} />`:l`<span part="initials">${this.#t}</span>`}
      </span>
    `}};kt([n({reflect:!0})],G.prototype,"src");kt([n()],G.prototype,"alt");kt([n()],G.prototype,"name");kt([n({reflect:!0})],G.prototype,"size");kt([H()],G.prototype,"_failed");b("mb-avatar",G);var ks=Object.defineProperty,Me=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&ks(t,e,s),s},_t=class extends p{constructor(){super(...arguments),this.size="md",this.label="Loading"}static{this.styles=[u,m`
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
    `}};Me([n({reflect:!0})],_t.prototype,"size");Me([n()],_t.prototype,"label");b("mb-spinner",_t);var se=class extends p{static{this.styles=[u,m`
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
    `}};b("mb-toolbar",se);var _s=Object.defineProperty,y=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&_s(t,e,s),s},Es="(max-width: 36rem)";function Ss(r){if(!r)return[];try{let t=JSON.parse(r);return Array.isArray(t)?t.filter(e=>!!e&&typeof e=="object"&&typeof e.id=="string"&&typeof e.label=="string").map(e=>({id:e.id,label:e.label,collapsed:!!e.collapsed,meta:typeof e.meta=="string"?e.meta:void 0,count:e.count===!1?!1:void 0})):[]}catch{return[]}}function zs(r,t){let e=Number(r),i=Number(t);return r!==""&&t!==""&&!Number.isNaN(e)&&!Number.isNaN(i)?e-i:r.localeCompare(t,void 0,{sensitivity:"base",numeric:!0})}var S=class extends p{constructor(){super(...arguments),this.label="",this.columns="",this.density="default",this.layout="auto",this.sections=[],this.sortKey="",this.sortDirection="asc",this.reorderable=!1,this.reorderLabel="Drag to reorder",this.sortLabel="Sort by {name}",this.hideCount=!1,this.stickyHeader=!1,this._sectionCounts={},this.#e=()=>this.#b(),this.#s=!1,this.#r=!1,this.#o=!1,this.#i=null,this.#l=null,this.#a="before",this.#n=null,this.#x=t=>{if(!this.#i)return;let e=document.elementsFromPoint(t.clientX,t.clientY),i=e.find(o=>o instanceof HTMLElement&&o.localName==="mb-table-row"&&o!==this.#i&&o.slot!=="head"&&!o.hasAttribute("head")),s=e.find(o=>o instanceof HTMLElement&&o.hasAttribute("data-section"));if(this.#k(),i){let o=i.getBoundingClientRect(),a=t.clientY<o.top+o.height/2?"before":"after";this.#l=i,this.#a=a,this.#n=i.section.trim()||null,i.setAttribute("data-drop",a);return}if(s){let o=s.getAttribute("data-section");o&&(this.#n=o,s.toggleAttribute("data-drop-section",!0))}},this.#u=()=>{let t=this.#i,e=this.#l,i=this.#a,s=this.#n;if(window.removeEventListener("pointermove",this.#x),window.removeEventListener("pointerup",this.#u),window.removeEventListener("pointercancel",this.#u),t?.removeAttribute("data-dragging"),this.#k(),this.#i=null,!!t){if(e){this.moveRow(t,{before:i==="before"?e:void 0,after:i==="after"?e:void 0,section:e.section.trim()||void 0});return}s!=null&&this.moveRow(t,{section:s})}},this.#f=()=>{this.#s||this.#r||queueMicrotask(()=>{this.#s||this.#r||(this.#b(),this.#y())})},this.#_=t=>{let e=t.composedPath(),i=e.find(s=>s instanceof HTMLElement&&s.localName==="mb-table-cell");!i?.sortKey.trim()||t.defaultPrevented||e.some(s=>s instanceof Element&&s.matches("mb-input, mb-select, mb-textarea, mb-button, a, input, select, textarea"))||this.#S(i.sortKey.trim())}}static{this.styles=[u,m`
      :host {
        display: block;
        inline-size: 100%;
        --mb-table-template: repeat(var(--mb-table-col-count, 1), minmax(0, 1fr));
      }

      .root {
        display: flex;
        flex-direction: column;
        gap: var(--mb-space-3);
      }

      .caption {
        margin: 0;
        font-family: var(--mb-font-display);
        font-size: var(--mb-font-size-lg);
        font-weight: 650;
        line-height: var(--mb-line-height-tight);
      }

      .frame {
        display: flex;
        flex-direction: column;
        gap: var(--mb-space-3);
        min-inline-size: 0;
      }

      .head {
        display: none;
      }

      .body {
        display: flex;
        flex-direction: column;
        gap: var(--mb-space-3);
        min-inline-size: 0;
      }

      .section {
        display: flex;
        flex-direction: column;
        gap: var(--mb-space-2);
        min-inline-size: 0;
      }

      .section-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: var(--mb-space-3);
        inline-size: 100%;
        margin: 0;
        padding-block: var(--mb-space-2);
        padding-inline: var(--mb-space-3);
        border: 1px solid var(--mb-color-border);
        border-radius: var(--mb-radius-md);
        background: var(--mb-color-bg);
        color: var(--mb-color-fg);
        font: inherit;
        font-family: var(--mb-font-display);
        font-weight: 650;
        text-align: start;
        cursor: pointer;
      }

      .section-head:focus-visible {
        outline: var(--mb-focus-ring);
        outline-offset: var(--mb-focus-offset);
      }

      .section-label {
        min-inline-size: 0;
      }

      .section-meta {
        display: inline-flex;
        align-items: center;
        gap: var(--mb-space-2);
        color: var(--mb-color-muted);
        font-family: var(--mb-font-body);
        font-size: var(--mb-font-size-sm);
        font-weight: 600;
      }

      .section-chevron {
        display: inline-block;
        transition: transform var(--mb-transition);
      }

      .section[data-collapsed] .section-chevron {
        transform: rotate(-90deg);
      }

      .section-rows {
        display: flex;
        flex-direction: column;
        gap: var(--mb-space-3);
        min-inline-size: 0;
      }

      .section[data-collapsed] .section-rows {
        display: none;
      }

      .ungrouped:not([data-has-content]) {
        display: none;
      }

      .empty:not([data-has-content]) {
        display: none;
      }

      :host([data-mode='table']) .frame {
        gap: 0;
        border: 1px solid var(--mb-color-border);
        border-radius: var(--mb-radius-lg);
        background: var(--mb-color-surface);
        overflow: clip;
      }

      /* Sticky headers need a non-clipping ancestor. */
      :host([sticky-header][data-mode='table']) .frame {
        overflow: visible;
      }

      :host([data-mode='table']) .head {
        display: block;
        background: var(--mb-color-bg);
        border-block-end: 1px solid var(--mb-color-border);
      }

      :host([sticky-header][data-mode='table']) .head {
        position: sticky;
        inset-block-start: 0;
        z-index: 2;
        background: var(--mb-color-bg);
      }

      :host([data-mode='table']) .body {
        gap: 0;
      }

      :host([data-mode='table']) .section {
        gap: 0;
      }

      :host([data-mode='table']) .section-head {
        border: none;
        border-radius: 0;
        border-block-end: 1px solid var(--mb-color-border);
        padding-inline: var(--mb-space-4);
      }

      :host([data-mode='table']) .section-rows {
        gap: 0;
      }

      :host([data-mode='table'][density='compact']) .root {
        gap: var(--mb-space-2);
      }

      :host([data-mode='table'][density='compact']) .section-head {
        padding-inline: var(--mb-space-3);
      }

      .section[data-drop-section] {
        outline: 2px solid var(--mb-color-accent);
        outline-offset: 2px;
        border-radius: var(--mb-radius-md);
      }
    `]}#t;#e;#s;#r;#o;#i;#l;#a;#n;connectedCallback(){super.connectedCallback(),this.setAttribute("role","table"),this.#t=window.matchMedia(Es),this.#t.addEventListener("change",this.#e),this.#p(),queueMicrotask(()=>this.#b())}disconnectedCallback(){this.#t?.removeEventListener("change",this.#e),super.disconnectedCallback()}updated(t){this.label?this.setAttribute("aria-label",this.label):this.removeAttribute("aria-label");let e=t.has("columns")||t.has("layout")||t.has("density")||t.has("sections")||t.has("reorderable")||t.has("reorderLabel")||t.has("sortLabel"),i=t.has("sortKey")||t.has("sortDirection")||t.has("sections");(e||i)&&queueMicrotask(()=>{e&&(this.#p(),this.#b()),i&&(this.#g(),this.#y())})}get#d(){return this.layout==="table"?"table":this.layout==="cards"||this.#t?.matches?"cards":"table"}get#c(){return this.sections.length>0}#p(){let t=this.columns.trim();if(!t){this.style.removeProperty("--mb-table-template"),this.style.removeProperty("--mb-table-col-count");return}if(/^\d+$/.test(t)){this.style.setProperty("--mb-table-col-count",t),this.style.setProperty("--mb-table-template",`repeat(${t}, minmax(0, 1fr))`);return}this.style.setProperty("--mb-table-template",t)}refreshRows(){this.#o||(this.#b(),this.#y())}#h(){return[...this.querySelectorAll("mb-table-row")].filter(t=>t.slot!=="head"&&!t.hasAttribute("head"))}#m(){return this.querySelector('mb-table-row[slot="head"]')??this.querySelector("mb-table-row[head]")}#v(){let t=new Set(this.sections.map(s=>s.id)),e={};for(let s of this.sections)e[s.id]=0;for(let s of this.#h()){let o=s.section.trim();if(o&&t.has(o)){let a=`section-${o}`;s.slot!==a&&(s.slot=a),e[o]=(e[o]??0)+1}else s.slot.startsWith("section-")&&(s.slot="")}let i=this._sectionCounts;Object.keys(e).length===Object.keys(i).length&&Object.keys(e).every(s=>i[s]===e[s])||(this._sectionCounts=e)}#b(){if(!this.#r){this.#r=!0;try{let t=this.#d;this.setAttribute("data-mode",t),this.#v(),this.querySelectorAll("mb-table-row").forEach(e=>{e.setAttribute("data-mode",t),e.toggleAttribute("data-compact",this.density==="compact");let i=e.slot==="head"||e.hasAttribute("head");e.toggleAttribute("data-reorderable",this.reorderable&&!i),e.toggleAttribute("data-reorder-spacer",this.reorderable&&i),this.reorderable&&!i?e.setAttribute("data-reorder-label",this.reorderLabel):e.removeAttribute("data-reorder-label")}),this.querySelectorAll("mb-table-cell").forEach(e=>{e.setAttribute("data-mode",t),e.toggleAttribute("data-compact",this.density==="compact"),(e.sortKey.trim()||e.sortable)&&e.setAttribute("data-sort-label",this.sortLabel)}),this.#E(),this.#g(),this.#$()}finally{this.#r=!1}}}#$(){if(!this.#c)return;let t=this.renderRoot.querySelector("slot.ungrouped-slot"),e=this.renderRoot.querySelector(".ungrouped");if(!t||!e)return;let i=t.assignedElements({flatten:!0}).some(s=>s.localName==="mb-table-row");e.toggleAttribute("data-has-content",i)}#E(){let t=this.#m();if(!t)return;let e=[...t.querySelectorAll("mb-table-cell")],i=e.map(s=>s.hideLabel||s.actions?"":(s.textContent??"").replace(/\s+/g," ").trim());if(e.length){this.columns.trim()||(this.style.setProperty("--mb-table-col-count",String(e.length)),this.style.setProperty("--mb-table-template",`repeat(${e.length}, minmax(0, 1fr))`));for(let s of this.#h())[...s.querySelectorAll(":scope > mb-table-cell")].forEach((o,a)=>{if(o.hideLabel||o.actions){o.dataset.labelLocked="true",o.label&&(o.label="");return}if(o.dataset.labelLocked==="true")return;if(o.hasAttribute("label")){o.dataset.labelLocked="true";return}let h=i[a];h&&o.label!==h&&(o.label=h)})}}#g(){let t=this.#m();if(t)for(let e of t.querySelectorAll("mb-table-cell")){let i=e.sortKey.trim(),s=!!i&&i===this.sortKey,o=s?this.sortDirection:null;i&&!e.sortable&&(e.sortable=!0),e.sortActive!==s&&(e.sortActive=s),e.sortDirection!==o&&(e.sortDirection=o)}}#A(t){let e=this.#m();return e?[...e.querySelectorAll("mb-table-cell")].findIndex(i=>i.sortKey.trim()===t):-1}#w(t,e){if(t.sortValue.trim()&&(!e||this.#A(e)<0))return t.sortValue.trim();let i=this.#A(e);if(i<0)return t.sortValue.trim();let s=t.querySelectorAll(":scope > mb-table-cell")[i];if(!s)return"";if(s.sortValue.trim())return s.sortValue.trim();let o=s.querySelector("mb-input, mb-select, mb-textarea, input, select, textarea");return o&&typeof o.value=="string"&&o.value!==""?o.value:(s.textContent??"").replace(/\s+/g," ").trim()}#y(){if(!(this.#s||!this.sortKey.trim())){this.#s=!0;try{let t=this.sortKey.trim(),e=this.sortDirection==="desc"?-1:1,i=this.#c?this.sections.map(s=>this.#h().filter(o=>o.section.trim()===s.id)):[this.#h()];for(let s of i){let o=[...s].sort((a,h)=>e*zs(this.#w(a,t),this.#w(h,t)));if(!(s.length===o.length&&s.every((a,h)=>a===o[h])))for(let a of o)this.appendChild(a)}}finally{this.#s=!1}}}#S(t){this.sortKey===t?this.sortDirection=this.sortDirection==="asc"?"desc":"asc":(this.sortKey=t,this.sortDirection="asc"),this.#g(),this.#y(),this.dispatchEvent(new CustomEvent("mb-sort",{detail:{key:this.sortKey,direction:this.sortDirection},bubbles:!0,composed:!0}))}beginReorder(t,e){!this.reorderable||t.head||t.slot==="head"||this.#i||e.button!==void 0&&e.button!==0||(e.preventDefault(),e.stopPropagation(),this.#i=t,t.toggleAttribute("data-dragging",!0),window.addEventListener("pointermove",this.#x),window.addEventListener("pointerup",this.#u),window.addEventListener("pointercancel",this.#u))}#k(){this.querySelectorAll("mb-table-row[data-drop]").forEach(t=>{t.removeAttribute("data-drop")}),this.renderRoot.querySelectorAll("[data-drop-section]").forEach(t=>{t.removeAttribute("data-drop-section")}),this.#l=null,this.#n=null}#x;#u;moveRow(t,e={}){if(t.head||t.slot==="head"||!this.contains(t))return;let i=t.section.trim(),s=e.section??i,o=!1;this.#o=!0,this.sortKey&&(this.sortKey="",this.#g());try{if(e.section!=null&&e.section!==i&&(t.section=e.section,s=e.section,o=!0),e.before&&e.before!==t)e.before.previousElementSibling!==t&&(e.before.before(t),o=!0),s=e.before.section.trim()||s,t.section.trim()!==s&&(t.section=s,o=!0);else if(e.after&&e.after!==t)e.after.nextElementSibling!==t&&(e.after.after(t),o=!0),s=e.after.section.trim()||s,t.section.trim()!==s&&(t.section=s,o=!0);else if(e.section!=null){let h=this.#h().filter(v=>v!==t&&v.section.trim()===e.section),d=h[h.length-1];d?(d.after(t),o=!0):(this.appendChild(t),o=!0)}}finally{this.#o=!1}if(!o)return;this.#b();let a=this.#h().map(h=>({id:h.id||h.getAttribute("data-id")||"",section:h.section.trim()}));this.dispatchEvent(new CustomEvent("mb-reorder",{detail:{rowId:t.id||t.getAttribute("data-id")||"",fromSection:i,toSection:s,beforeId:e.before?e.before.id||e.before.getAttribute("data-id")||"":null,afterId:e.after?e.after.id||e.after.getAttribute("data-id")||"":null,order:a},bubbles:!0,composed:!0}))}#z(t){let e=this.sections.findIndex(o=>o.id===t);if(e<0)return;let i=this.sections.map((o,a)=>a===e?{...o,collapsed:!o.collapsed}:o);this.sections=i;let s=!!i[e]?.collapsed;this.dispatchEvent(new CustomEvent("mb-section-toggle",{detail:{id:t,collapsed:s},bubbles:!0,composed:!0}))}#f;#C(t){let e=t.target.assignedNodes({flatten:!0}).length>0;this.renderRoot.querySelector(".empty")?.toggleAttribute("data-has-content",e)}#L(t){return!(this.hideCount||t.count===!1)}#_;render(){return l`
      <div part="root" class="root">
        ${this.label?l`<div part="caption" class="caption">${this.label}</div>`:c}
        <div part="frame" class="frame">
          <div part="head" class="head" @click=${this.#_}>
            <slot name="head" @slotchange=${this.#f}></slot>
          </div>
          <div part="body" class="body">
            ${this.#c?l`
                  ${ct(this.sections,t=>t.id,t=>l`
                      <section
                        part="section"
                        class="section"
                        data-section=${t.id}
                        ?data-collapsed=${!!t.collapsed}
                      >
                        <button
                          type="button"
                          part="section-head"
                          class="section-head"
                          aria-expanded=${t.collapsed?"false":"true"}
                          @click=${()=>this.#z(t.id)}
                        >
                          <span class="section-label">${t.label}</span>
                          <span class="section-meta">
                            ${t.meta?l`<span part="section-custom-meta">${t.meta}</span>`:c}
                            <slot name=${`section-meta-${t.id}`}></slot>
                            ${this.#L(t)?l`<span part="section-count"
                                  >${this._sectionCounts[t.id]??0}</span
                                >`:c}
                            <span class="section-chevron" aria-hidden="true">▾</span>
                          </span>
                        </button>
                        <div part="section-rows" class="section-rows">
                          <slot
                            name=${`section-${t.id}`}
                            @slotchange=${this.#f}
                          ></slot>
                        </div>
                      </section>
                    `)}
                  <div part="ungrouped" class="ungrouped section">
                    <slot
                      class="ungrouped-slot"
                      @slotchange=${this.#f}
                    ></slot>
                  </div>
                `:l`<slot @slotchange=${this.#f}></slot>`}
          </div>
        </div>
        <div part="empty" class="empty">
          <slot name="empty" @slotchange=${this.#C}></slot>
        </div>
      </div>
    `}};y([n()],S.prototype,"label");y([n()],S.prototype,"columns");y([n({reflect:!0})],S.prototype,"density");y([n({reflect:!0})],S.prototype,"layout");y([n({attribute:"sections",converter:{fromAttribute:Ss,toAttribute(r){return r?.length?JSON.stringify(r):null}}})],S.prototype,"sections");y([n({attribute:"sort-key",reflect:!0})],S.prototype,"sortKey");y([n({attribute:"sort-direction",reflect:!0})],S.prototype,"sortDirection");y([n({type:Boolean,reflect:!0})],S.prototype,"reorderable");y([n({attribute:"reorder-label"})],S.prototype,"reorderLabel");y([n({attribute:"sort-label"})],S.prototype,"sortLabel");y([n({type:Boolean,reflect:!0,attribute:"hide-count"})],S.prototype,"hideCount");y([n({type:Boolean,reflect:!0,attribute:"sticky-header"})],S.prototype,"stickyHeader");y([H()],S.prototype,"_sectionCounts");var dt=class extends p{constructor(){super(...arguments),this.head=!1,this.section="",this.sortValue="",this.#t=t=>{this.closest("mb-table")?.beginReorder(this,t)}}static{this.styles=[u,m`
      :host {
        display: block;
        min-inline-size: 0;
      }

      .wrap {
        display: flex;
        align-items: stretch;
        gap: var(--mb-space-2);
        min-inline-size: 0;
      }

      .handle,
      .spacer {
        flex: none;
        inline-size: 1.25rem;
        align-self: center;
      }

      .spacer {
        visibility: hidden;
      }

      .handle {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        margin: 0;
        padding: 0;
        border: none;
        border-radius: var(--mb-radius-sm);
        background: transparent;
        color: var(--mb-color-muted);
        font: inherit;
        line-height: 1;
        cursor: grab;
        touch-action: none;
        user-select: none;
      }

      .handle:focus-visible {
        outline: var(--mb-focus-ring);
        outline-offset: var(--mb-focus-offset);
      }

      .handle:active {
        cursor: grabbing;
      }

      :host(:not([data-reorderable]):not([data-reorder-spacer])) .handle,
      :host(:not([data-reorderable]):not([data-reorder-spacer])) .spacer {
        display: none;
      }

      :host([data-dragging]) {
        opacity: 0.45;
      }

      :host([data-drop='before']) {
        box-shadow: inset 0 2px 0 var(--mb-color-accent);
      }

      :host([data-drop='after']) {
        box-shadow: inset 0 -2px 0 var(--mb-color-accent);
      }

      .row {
        display: grid;
        grid-template-columns: var(--mb-table-template);
        align-items: center;
        gap: var(--mb-space-3);
        min-inline-size: 0;
        flex: 1;
      }

      :host([data-mode='table']) .wrap {
        padding-block: var(--mb-space-3);
        padding-inline: var(--mb-space-4);
        border-block-end: 1px solid var(--mb-color-border);
        background: var(--mb-color-surface);
      }

      :host([data-mode='table'][data-compact]) .wrap {
        padding-block: var(--mb-space-2);
        padding-inline: var(--mb-space-3);
      }

      :host([data-mode='table'][data-compact]) .row {
        gap: var(--mb-space-2);
      }

      :host([data-mode='table']:last-of-type) .wrap,
      :host([data-mode='table'][slot='head']) .wrap {
        border-block-end: none;
      }

      :host([slot='head']) .wrap,
      :host([head]) .wrap {
        font-size: var(--mb-font-size-sm);
        font-weight: 650;
        color: var(--mb-color-muted);
        background: transparent;
        padding-block: var(--mb-space-2);
      }

      :host([data-mode='cards']) .wrap {
        padding-block: var(--mb-space-4);
        padding-inline: var(--mb-space-4);
        background: var(--mb-color-surface);
        border: 1px solid var(--mb-color-border);
        border-radius: var(--mb-radius-lg);
      }

      :host([data-mode='cards']) .row {
        display: flex;
        flex-direction: column;
        align-items: stretch;
        gap: var(--mb-space-3);
      }

      :host([data-mode='cards'][data-compact]) .wrap {
        padding-block: var(--mb-space-3);
        padding-inline: var(--mb-space-3);
      }

      :host([data-mode='cards'][data-compact]) .row {
        gap: var(--mb-space-2);
      }

      :host([data-mode='cards'][slot='head']),
      :host([data-mode='cards'][head]) {
        display: none;
      }

      :host([data-mode='cards'][data-reorderable]) .handle {
        align-self: flex-start;
        margin-block-start: 0.15rem;
      }
    `]}connectedCallback(){super.connectedCallback(),this.setAttribute("role","row"),this.head&&this.slot!=="head"&&(this.slot="head")}attributeChangedCallback(t,e,i){super.attributeChangedCallback(t,e,i),t==="data-reorder-label"&&e!==i&&this.requestUpdate()}updated(t){if(t.has("head")&&this.head&&(this.slot="head"),t.has("section")){let i=t.get("section");if(i!==void 0||this.section){let s=this.closest("mb-table");s&&i!==void 0&&s.refreshRows()}}let e=(this.slot==="head"||this.head)&&this.getAttribute("data-mode")==="cards";this.toggleAttribute("aria-hidden",e)}#t;#e(){return this.getAttribute("data-reorder-label")?.trim()||this.closest("mb-table")?.reorderLabel?.trim()||"Drag to reorder"}render(){let t=this.hasAttribute("data-reorderable"),e=this.hasAttribute("data-reorder-spacer");return l`
      <div part="wrap" class="wrap">
        ${t?l`
              <button
                type="button"
                part="handle"
                class="handle"
                aria-label=${this.#e()}
                @pointerdown=${this.#t}
              >
                ⠿
              </button>
            `:e?l`<span class="spacer" aria-hidden="true"></span>`:c}
        <div part="row" class="row">
          <slot></slot>
        </div>
      </div>
    `}};y([n({type:Boolean,reflect:!0})],dt.prototype,"head");y([n({reflect:!0})],dt.prototype,"section");y([n({attribute:"sort-value"})],dt.prototype,"sortValue");var P=class extends p{constructor(){super(...arguments),this.label="",this.align="start",this.primary=!1,this.hideLabel=!1,this.actions=!1,this.sortKey="",this.sortable=!1,this.sortValue="",this.sortActive=!1,this.sortDirection=null}static{this.styles=[u,m`
      :host {
        display: block;
        min-inline-size: 0;
      }

      .cell {
        display: flex;
        flex-direction: column;
        align-items: stretch;
        gap: var(--mb-space-1);
        min-inline-size: 0;
      }

      .label {
        display: none;
        font-size: var(--mb-font-size-sm);
        font-weight: 600;
        color: var(--mb-color-muted);
      }

      .value {
        min-inline-size: 0;
        max-inline-size: 100%;
      }

      .sort {
        display: inline-flex;
        align-items: center;
        gap: var(--mb-space-1);
        max-inline-size: 100%;
        margin: 0;
        padding: 0;
        border: none;
        background: transparent;
        color: inherit;
        font: inherit;
        font-weight: inherit;
        text-align: inherit;
        cursor: pointer;
      }

      .sort:focus-visible {
        outline: var(--mb-focus-ring);
        outline-offset: var(--mb-focus-offset);
      }

      .sort-indicator {
        color: var(--mb-color-muted);
        font-size: 0.75em;
      }

      :host([sort-active]) .sort-indicator {
        color: var(--mb-color-accent);
      }

      :host([align='center']) .cell {
        align-items: center;
        text-align: center;
      }

      :host([align='end']) .cell {
        align-items: end;
        text-align: end;
      }

      :host([data-mode='cards']) .label:not([hidden]) {
        display: block;
      }

      :host([data-mode='cards'][primary]) .value {
        font-family: var(--mb-font-display);
        font-weight: 650;
        font-size: var(--mb-font-size-md);
      }

      :host([data-mode='cards'][align='end']) .cell,
      :host([data-mode='cards'][align='center']) .cell {
        align-items: stretch;
        text-align: start;
      }

      :host([data-mode='cards'][align='end']) .value {
        display: flex;
        justify-content: flex-end;
        flex-wrap: wrap;
        gap: var(--mb-space-2);
      }

      :host([actions][data-mode='cards']) .value,
      :host([actions][data-mode='table']) .value {
        display: flex;
        justify-content: flex-end;
        flex-wrap: wrap;
        align-items: center;
        gap: var(--mb-space-2);
      }
    `]}connectedCallback(){super.connectedCallback(),this.#e()}attributeChangedCallback(t,e,i){super.attributeChangedCallback(t,e,i),t==="data-sort-label"&&e!==i&&this.requestUpdate()}updated(t){if(this.#e(),t.has("sortKey")&&this.sortKey.trim()&&(this.sortable=!0),(t.has("hideLabel")||t.has("actions"))&&(this.hideLabel||this.actions)&&(this.dataset.labelLocked="true",this.label&&(this.label="")),this.#t()&&this.sortKey.trim()){let e=this.sortActive&&this.sortDirection?this.sortDirection:"none";this.setAttribute("aria-sort",e)}else this.removeAttribute("aria-sort")}#t(){let t=this.parentElement;return t?.slot==="head"||t?.hasAttribute("head")===!0}#e(){this.setAttribute("role",this.#t()?"columnheader":"cell")}#s(){return!this.sortActive||!this.sortDirection?"\u2195":this.sortDirection==="asc"?"\u2191":"\u2193"}#r(){let t=this.sortKey.trim()||this.label.trim()||(this.textContent??"").replace(/\s+/g," ").trim()||"column",e=this.getAttribute("data-sort-label")?.trim()||this.closest("mb-table")?.sortLabel?.trim()||"Sort by {name}";return e.includes("{name}")?e.replace(/\{name\}/g,t):`${e} ${t}`.trim()}render(){let t=!!this.label&&this.getAttribute("data-mode")==="cards"&&!this.#t()&&!this.hideLabel&&!this.actions,e=this.#t()&&(this.sortable||!!this.sortKey.trim());return l`
      <div part="cell" class="cell">
        <span part="label" class="label" ?hidden=${!t}>${this.label}</span>
        <div part="value" class="value">
          ${e?l`
                <button
                  type="button"
                  part="sort"
                  class="sort"
                  aria-label=${this.#r()}
                >
                  <slot></slot>
                  <span class="sort-indicator" aria-hidden="true">${this.#s()}</span>
                </button>
              `:l`<slot></slot>`}
        </div>
      </div>
    `}};y([n()],P.prototype,"label");y([n({reflect:!0})],P.prototype,"align");y([n({type:Boolean,reflect:!0})],P.prototype,"primary");y([n({type:Boolean,reflect:!0,attribute:"hide-label"})],P.prototype,"hideLabel");y([n({type:Boolean,reflect:!0})],P.prototype,"actions");y([n({attribute:"sort-key",reflect:!0})],P.prototype,"sortKey");y([n({type:Boolean,reflect:!0})],P.prototype,"sortable");y([n({attribute:"sort-value"})],P.prototype,"sortValue");y([n({type:Boolean,reflect:!0,attribute:"sort-active"})],P.prototype,"sortActive");y([n({attribute:!1})],P.prototype,"sortDirection");b("mb-table",S);b("mb-table-row",dt);b("mb-table-cell",P);
