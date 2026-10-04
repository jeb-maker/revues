var Lt=globalThis,Dt=Lt.ShadowRoot&&(Lt.ShadyCSS===void 0||Lt.ShadyCSS.nativeShadow)&&"adoptedStyleSheets"in Document.prototype&&"replace"in CSSStyleSheet.prototype,Qt=Symbol(),me=new WeakMap,vt=class{constructor(t,e,i){if(this._$cssResult$=!0,i!==Qt)throw Error("CSSResult is not constructable. Use `unsafeCSS` or `css` instead.");this.cssText=t,this.t=e}get styleSheet(){let t=this.o,e=this.t;if(Dt&&t===void 0){let i=e!==void 0&&e.length===1;i&&(t=me.get(e)),t===void 0&&((this.o=t=new CSSStyleSheet).replaceSync(this.cssText),i&&me.set(e,t))}return t}toString(){return this.cssText}},be=r=>new vt(typeof r=="string"?r:r+"",void 0,Qt),m=(r,...t)=>{let e=r.length===1?r[0]:t.reduce((i,s,o)=>i+(a=>{if(a._$cssResult$===!0)return a.cssText;if(typeof a=="number")return a;throw Error("Value passed to 'css' function must be a 'css' function result: "+a+". Use 'unsafeCSS' to pass non-literal values, but take care to ensure page security.")})(s)+r[o+1],r[0]);return new vt(e,r,Qt)},ue=(r,t)=>{if(Dt)r.adoptedStyleSheets=t.map(e=>e instanceof CSSStyleSheet?e:e.styleSheet);else for(let e of t){let i=document.createElement("style"),s=Lt.litNonce;s!==void 0&&i.setAttribute("nonce",s),i.textContent=e.cssText,r.appendChild(i)}},Yt=Dt?r=>r:r=>r instanceof CSSStyleSheet?(t=>{let e="";for(let i of t.cssRules)e+=i.cssText;return be(e)})(r):r;var{is:We,defineProperty:Qe,getOwnPropertyDescriptor:Ye,getOwnPropertyNames:Ge,getOwnPropertySymbols:Xe,getPrototypeOf:Ze}=Object,Pt=globalThis,fe=Pt.trustedTypes,ts=fe?fe.emptyScript:"",es=Pt.reactiveElementPolyfillSupport,gt=(r,t)=>r,yt={toAttribute(r,t){switch(t){case Boolean:r=r?ts:null;break;case Object:case Array:r=r==null?r:JSON.stringify(r)}return r},fromAttribute(r,t){let e=r;switch(t){case Boolean:e=r!==null;break;case Number:e=r===null?null:Number(r);break;case Object:case Array:try{e=JSON.parse(r)}catch{e=null}}return e}},Ot=(r,t)=>!We(r,t),ve={attribute:!0,type:String,converter:yt,reflect:!1,useDefault:!1,hasChanged:Ot};Symbol.metadata??=Symbol("metadata"),Pt.litPropertyMetadata??=new WeakMap;var J=class extends HTMLElement{static addInitializer(t){this._$Ei(),(this.l??=[]).push(t)}static get observedAttributes(){return this.finalize(),this._$Eh&&[...this._$Eh.keys()]}static createProperty(t,e=ve){if(e.state&&(e.attribute=!1),this._$Ei(),this.prototype.hasOwnProperty(t)&&((e=Object.create(e)).wrapped=!0),this.elementProperties.set(t,e),!e.noAccessor){let i=Symbol(),s=this.getPropertyDescriptor(t,i,e);s!==void 0&&Qe(this.prototype,t,s)}}static getPropertyDescriptor(t,e,i){let{get:s,set:o}=Ye(this.prototype,t)??{get(){return this[e]},set(a){this[e]=a}};return{get:s,set(a){let d=s?.call(this);o?.call(this,a),this.requestUpdate(t,d,i)},configurable:!0,enumerable:!0}}static getPropertyOptions(t){return this.elementProperties.get(t)??ve}static _$Ei(){if(this.hasOwnProperty(gt("elementProperties")))return;let t=Ze(this);t.finalize(),t.l!==void 0&&(this.l=[...t.l]),this.elementProperties=new Map(t.elementProperties)}static finalize(){if(this.hasOwnProperty(gt("finalized")))return;if(this.finalized=!0,this._$Ei(),this.hasOwnProperty(gt("properties"))){let e=this.properties,i=[...Ge(e),...Xe(e)];for(let s of i)this.createProperty(s,e[s])}let t=this[Symbol.metadata];if(t!==null){let e=litPropertyMetadata.get(t);if(e!==void 0)for(let[i,s]of e)this.elementProperties.set(i,s)}this._$Eh=new Map;for(let[e,i]of this.elementProperties){let s=this._$Eu(e,i);s!==void 0&&this._$Eh.set(s,e)}this.elementStyles=this.finalizeStyles(this.styles)}static finalizeStyles(t){let e=[];if(Array.isArray(t)){let i=new Set(t.flat(1/0).reverse());for(let s of i)e.unshift(Yt(s))}else t!==void 0&&e.push(Yt(t));return e}static _$Eu(t,e){let i=e.attribute;return i===!1?void 0:typeof i=="string"?i:typeof t=="string"?t.toLowerCase():void 0}constructor(){super(),this._$Ep=void 0,this.isUpdatePending=!1,this.hasUpdated=!1,this._$Em=null,this._$Ev()}_$Ev(){this._$ES=new Promise(t=>this.enableUpdating=t),this._$AL=new Map,this._$E_(),this.requestUpdate(),this.constructor.l?.forEach(t=>t(this))}addController(t){(this._$EO??=new Set).add(t),this.renderRoot!==void 0&&this.isConnected&&t.hostConnected?.()}removeController(t){this._$EO?.delete(t)}_$E_(){let t=new Map,e=this.constructor.elementProperties;for(let i of e.keys())this.hasOwnProperty(i)&&(t.set(i,this[i]),delete this[i]);t.size>0&&(this._$Ep=t)}createRenderRoot(){let t=this.shadowRoot??this.attachShadow(this.constructor.shadowRootOptions);return ue(t,this.constructor.elementStyles),t}connectedCallback(){this.renderRoot??=this.createRenderRoot(),this.enableUpdating(!0),this._$EO?.forEach(t=>t.hostConnected?.())}enableUpdating(t){}disconnectedCallback(){this._$EO?.forEach(t=>t.hostDisconnected?.())}attributeChangedCallback(t,e,i){this._$AK(t,i)}_$ET(t,e){let i=this.constructor.elementProperties.get(t),s=this.constructor._$Eu(t,i);if(s!==void 0&&i.reflect===!0){let o=(i.converter?.toAttribute!==void 0?i.converter:yt).toAttribute(e,i.type);this._$Em=t,o==null?this.removeAttribute(s):this.setAttribute(s,o),this._$Em=null}}_$AK(t,e){let i=this.constructor,s=i._$Eh.get(t);if(s!==void 0&&this._$Em!==s){let o=i.getPropertyOptions(s),a=typeof o.converter=="function"?{fromAttribute:o.converter}:o.converter?.fromAttribute!==void 0?o.converter:yt;this._$Em=s;let d=a.fromAttribute(e,o.type);this[s]=d??this._$Ej?.get(s)??d,this._$Em=null}}requestUpdate(t,e,i,s=!1,o){if(t!==void 0){let a=this.constructor;if(s===!1&&(o=this[t]),i??=a.getPropertyOptions(t),!((i.hasChanged??Ot)(o,e)||i.useDefault&&i.reflect&&o===this._$Ej?.get(t)&&!this.hasAttribute(a._$Eu(t,i))))return;this.C(t,e,i)}this.isUpdatePending===!1&&(this._$ES=this._$EP())}C(t,e,{useDefault:i,reflect:s,wrapped:o},a){i&&!(this._$Ej??=new Map).has(t)&&(this._$Ej.set(t,a??e??this[t]),o!==!0||a!==void 0)||(this._$AL.has(t)||(this.hasUpdated||i||(e=void 0),this._$AL.set(t,e)),s===!0&&this._$Em!==t&&(this._$Eq??=new Set).add(t))}async _$EP(){this.isUpdatePending=!0;try{await this._$ES}catch(e){Promise.reject(e)}let t=this.scheduleUpdate();return t!=null&&await t,!this.isUpdatePending}scheduleUpdate(){return this.performUpdate()}performUpdate(){if(!this.isUpdatePending)return;if(!this.hasUpdated){if(this.renderRoot??=this.createRenderRoot(),this._$Ep){for(let[s,o]of this._$Ep)this[s]=o;this._$Ep=void 0}let i=this.constructor.elementProperties;if(i.size>0)for(let[s,o]of i){let{wrapped:a}=o,d=this[s];a!==!0||this._$AL.has(s)||d===void 0||this.C(s,void 0,o,d)}}let t=!1,e=this._$AL;try{t=this.shouldUpdate(e),t?(this.willUpdate(e),this._$EO?.forEach(i=>i.hostUpdate?.()),this.update(e)):this._$EM()}catch(i){throw t=!1,this._$EM(),i}t&&this._$AE(e)}willUpdate(t){}_$AE(t){this._$EO?.forEach(e=>e.hostUpdated?.()),this.hasUpdated||(this.hasUpdated=!0,this.firstUpdated(t)),this.updated(t)}_$EM(){this._$AL=new Map,this.isUpdatePending=!1}get updateComplete(){return this.getUpdateComplete()}getUpdateComplete(){return this._$ES}shouldUpdate(t){return!0}update(t){this._$Eq&&=this._$Eq.forEach(e=>this._$ET(e,this[e])),this._$EM()}updated(t){}firstUpdated(t){}};J.elementStyles=[],J.shadowRootOptions={mode:"open"},J[gt("elementProperties")]=new Map,J[gt("finalized")]=new Map,es?.({ReactiveElement:J}),(Pt.reactiveElementVersions??=[]).push("2.1.2");var Xt=globalThis,ge=r=>r,Rt=Xt.trustedTypes,ye=Rt?Rt.createPolicy("lit-html",{createHTML:r=>r}):void 0,Zt="$lit$",W=`lit$${Math.random().toFixed(9).slice(2)}$`,te="?"+W,ss=`<${te}>`,lt=document,xt=()=>lt.createComment(""),kt=r=>r===null||typeof r!="object"&&typeof r!="function",ee=Array.isArray,_e=r=>ee(r)||typeof r?.[Symbol.iterator]=="function",Gt=`[ 	
\f\r]`,$t=/<(?:(!--|\/[^a-zA-Z])|(\/?[a-zA-Z][^>\s]*)|(\/?$))/g,$e=/-->/g,xe=/>/g,at=RegExp(`>|${Gt}(?:([^\\s"'>=/]+)(${Gt}*=${Gt}*(?:[^ 	
\f\r"'\`<>=]|("|')|))|$)`,"g"),ke=/'/g,we=/"/g,Ce=/^(?:script|style|textarea|title)$/i,se=r=>(t,...e)=>({_$litType$:r,strings:t,values:e}),l=se(1),Ws=se(2),Qs=se(3),Q=Symbol.for("lit-noChange"),c=Symbol.for("lit-nothing"),Ae=new WeakMap,nt=lt.createTreeWalker(lt,129);function ze(r,t){if(!ee(r)||!r.hasOwnProperty("raw"))throw Error("invalid template strings array");return ye!==void 0?ye.createHTML(t):t}var Se=(r,t)=>{let e=r.length-1,i=[],s,o=t===2?"<svg>":t===3?"<math>":"",a=$t;for(let d=0;d<e;d++){let h=r[d],v,A,f=-1,x=0;for(;x<h.length&&(a.lastIndex=x,A=a.exec(h),A!==null);)x=a.lastIndex,a===$t?A[1]==="!--"?a=$e:A[1]!==void 0?a=xe:A[2]!==void 0?(Ce.test(A[2])&&(s=RegExp("</"+A[2],"g")),a=at):A[3]!==void 0&&(a=at):a===at?A[0]===">"?(a=s??$t,f=-1):A[1]===void 0?f=-2:(f=a.lastIndex-A[2].length,v=A[1],a=A[3]===void 0?at:A[3]==='"'?we:ke):a===we||a===ke?a=at:a===$e||a===xe?a=$t:(a=at,s=void 0);let g=a===at&&r[d+1].startsWith("/>")?" ":"";o+=a===$t?h+ss:f>=0?(i.push(v),h.slice(0,f)+Zt+h.slice(f)+W+g):h+W+(f===-2?d:g)}return[ze(r,o+(r[e]||"<?>")+(t===2?"</svg>":t===3?"</math>":"")),i]},wt=class r{constructor({strings:t,_$litType$:e},i){let s;this.parts=[];let o=0,a=0,d=t.length-1,h=this.parts,[v,A]=Se(t,e);if(this.el=r.createElement(v,i),nt.currentNode=this.el.content,e===2||e===3){let f=this.el.content.firstChild;f.replaceWith(...f.childNodes)}for(;(s=nt.nextNode())!==null&&h.length<d;){if(s.nodeType===1){if(s.hasAttributes())for(let f of s.getAttributeNames())if(f.endsWith(Zt)){let x=A[a++],g=s.getAttribute(f).split(W),C=/([.?@])?(.*)/.exec(x);h.push({type:1,index:o,name:C[2],strings:g,ctor:C[1]==="."?Tt:C[1]==="?"?Bt:C[1]==="@"?Vt:ht}),s.removeAttribute(f)}else f.startsWith(W)&&(h.push({type:6,index:o}),s.removeAttribute(f));if(Ce.test(s.tagName)){let f=s.textContent.split(W),x=f.length-1;if(x>0){s.textContent=Rt?Rt.emptyScript:"";for(let g=0;g<x;g++)s.append(f[g],xt()),nt.nextNode(),h.push({type:2,index:++o});s.append(f[x],xt())}}}else if(s.nodeType===8)if(s.data===te)h.push({type:2,index:o});else{let f=-1;for(;(f=s.data.indexOf(W,f+1))!==-1;)h.push({type:7,index:o}),f+=W.length-1}o++}}static createElement(t,e){let i=lt.createElement("template");return i.innerHTML=t,i}};function ct(r,t,e=r,i){if(t===Q)return t;let s=i!==void 0?e._$Co?.[i]:e._$Cl,o=kt(t)?void 0:t._$litDirective$;return s?.constructor!==o&&(s?._$AO?.(!1),o===void 0?s=void 0:(s=new o(r),s._$AT(r,e,i)),i!==void 0?(e._$Co??=[])[i]=s:e._$Cl=s),s!==void 0&&(t=ct(r,s._$AS(r,t.values),s,i)),t}var qt=class{constructor(t,e){this._$AV=[],this._$AN=void 0,this._$AD=t,this._$AM=e}get parentNode(){return this._$AM.parentNode}get _$AU(){return this._$AM._$AU}u(t){let{el:{content:e},parts:i}=this._$AD,s=(t?.creationScope??lt).importNode(e,!0);nt.currentNode=s;let o=nt.nextNode(),a=0,d=0,h=i[0];for(;h!==void 0;){if(a===h.index){let v;h.type===2?v=new mt(o,o.nextSibling,this,t):h.type===1?v=new h.ctor(o,h.name,h.strings,this,t):h.type===6&&(v=new Nt(o,this,t)),this._$AV.push(v),h=i[++d]}a!==h?.index&&(o=nt.nextNode(),a++)}return nt.currentNode=lt,s}p(t){let e=0;for(let i of this._$AV)i!==void 0&&(i.strings!==void 0?(i._$AI(t,i,e),e+=i.strings.length-2):i._$AI(t[e])),e++}},mt=class r{get _$AU(){return this._$AM?._$AU??this._$Cv}constructor(t,e,i,s){this.type=2,this._$AH=c,this._$AN=void 0,this._$AA=t,this._$AB=e,this._$AM=i,this.options=s,this._$Cv=s?.isConnected??!0}get parentNode(){let t=this._$AA.parentNode,e=this._$AM;return e!==void 0&&t?.nodeType===11&&(t=e.parentNode),t}get startNode(){return this._$AA}get endNode(){return this._$AB}_$AI(t,e=this){t=ct(this,t,e),kt(t)?t===c||t==null||t===""?(this._$AH!==c&&this._$AR(),this._$AH=c):t!==this._$AH&&t!==Q&&this._(t):t._$litType$!==void 0?this.$(t):t.nodeType!==void 0?this.T(t):_e(t)?this.k(t):this._(t)}O(t){return this._$AA.parentNode.insertBefore(t,this._$AB)}T(t){this._$AH!==t&&(this._$AR(),this._$AH=this.O(t))}_(t){this._$AH!==c&&kt(this._$AH)?this._$AA.nextSibling.data=t:this.T(lt.createTextNode(t)),this._$AH=t}$(t){let{values:e,_$litType$:i}=t,s=typeof i=="number"?this._$AC(t):(i.el===void 0&&(i.el=wt.createElement(ze(i.h,i.h[0]),this.options)),i);if(this._$AH?._$AD===s)this._$AH.p(e);else{let o=new qt(s,this),a=o.u(this.options);o.p(e),this.T(a),this._$AH=o}}_$AC(t){let e=Ae.get(t.strings);return e===void 0&&Ae.set(t.strings,e=new wt(t)),e}k(t){ee(this._$AH)||(this._$AH=[],this._$AR());let e=this._$AH,i,s=0;for(let o of t)s===e.length?e.push(i=new r(this.O(xt()),this.O(xt()),this,this.options)):i=e[s],i._$AI(o),s++;s<e.length&&(this._$AR(i&&i._$AB.nextSibling,s),e.length=s)}_$AR(t=this._$AA.nextSibling,e){for(this._$AP?.(!1,!0,e);t!==this._$AB;){let i=ge(t).nextSibling;ge(t).remove(),t=i}}setConnected(t){this._$AM===void 0&&(this._$Cv=t,this._$AP?.(t))}},ht=class{get tagName(){return this.element.tagName}get _$AU(){return this._$AM._$AU}constructor(t,e,i,s,o){this.type=1,this._$AH=c,this._$AN=void 0,this.element=t,this.name=e,this._$AM=s,this.options=o,i.length>2||i[0]!==""||i[1]!==""?(this._$AH=Array(i.length-1).fill(new String),this.strings=i):this._$AH=c}_$AI(t,e=this,i,s){let o=this.strings,a=!1;if(o===void 0)t=ct(this,t,e,0),a=!kt(t)||t!==this._$AH&&t!==Q,a&&(this._$AH=t);else{let d=t,h,v;for(t=o[0],h=0;h<o.length-1;h++)v=ct(this,d[i+h],e,h),v===Q&&(v=this._$AH[h]),a||=!kt(v)||v!==this._$AH[h],v===c?t=c:t!==c&&(t+=(v??"")+o[h+1]),this._$AH[h]=v}a&&!s&&this.j(t)}j(t){t===c?this.element.removeAttribute(this.name):this.element.setAttribute(this.name,t??"")}},Tt=class extends ht{constructor(){super(...arguments),this.type=3}j(t){this.element[this.name]=t===c?void 0:t}},Bt=class extends ht{constructor(){super(...arguments),this.type=4}j(t){this.element.toggleAttribute(this.name,!!t&&t!==c)}},Vt=class extends ht{constructor(t,e,i,s,o){super(t,e,i,s,o),this.type=5}_$AI(t,e=this){if((t=ct(this,t,e,0)??c)===Q)return;let i=this._$AH,s=t===c&&i!==c||t.capture!==i.capture||t.once!==i.once||t.passive!==i.passive,o=t!==c&&(i===c||s);s&&this.element.removeEventListener(this.name,this,i),o&&this.element.addEventListener(this.name,this,t),this._$AH=t}handleEvent(t){typeof this._$AH=="function"?this._$AH.call(this.options?.host??this.element,t):this._$AH.handleEvent(t)}},Nt=class{constructor(t,e,i){this.element=t,this.type=6,this._$AN=void 0,this._$AM=e,this.options=i}get _$AU(){return this._$AM._$AU}_$AI(t){ct(this,t)}},Ee={M:Zt,P:W,A:te,C:1,L:Se,R:qt,D:_e,V:ct,I:mt,H:ht,N:Bt,U:Vt,B:Tt,F:Nt},rs=Xt.litHtmlPolyfillSupport;rs?.(wt,mt),(Xt.litHtmlVersions??=[]).push("3.3.3");var Me=(r,t,e)=>{let i=e?.renderBefore??t,s=i._$litPart$;if(s===void 0){let o=e?.renderBefore??null;i._$litPart$=s=new mt(t.insertBefore(xt(),o),o,void 0,e??{})}return s._$AI(r),s};var re=globalThis,p=class extends J{constructor(){super(...arguments),this.renderOptions={host:this},this._$Do=void 0}createRenderRoot(){let t=super.createRenderRoot();return this.renderOptions.renderBefore??=t.firstChild,t}update(t){let e=this.render();this.hasUpdated||(this.renderOptions.isConnected=this.isConnected),super.update(t),this._$Do=Me(e,this.renderRoot,this.renderOptions)}connectedCallback(){super.connectedCallback(),this._$Do?.setConnected(!0)}disconnectedCallback(){super.disconnectedCallback(),this._$Do?.setConnected(!1)}render(){return Q}};p._$litElement$=!0,p.finalized=!0,re.litElementHydrateSupport?.({LitElement:p});var is=re.litElementPolyfillSupport;is?.({LitElement:p});(re.litElementVersions??=[]).push("4.2.2");var os={attribute:!0,type:String,converter:yt,reflect:!1,hasChanged:Ot},as=(r=os,t,e)=>{let{kind:i,metadata:s}=e,o=globalThis.litPropertyMetadata.get(s);if(o===void 0&&globalThis.litPropertyMetadata.set(s,o=new Map),i==="setter"&&((r=Object.create(r)).wrapped=!0),o.set(e.name,r),i==="accessor"){let{name:a}=e;return{set(d){let h=t.get.call(this);t.set.call(this,d),this.requestUpdate(a,h,r,!0,d)},init(d){return d!==void 0&&this.C(a,void 0,r,d),d}}}if(i==="setter"){let{name:a}=e;return function(d){let h=this[a];t.call(this,d),this.requestUpdate(a,h,r,!0,d)}}throw Error("Unsupported decorator location: "+i)};function n(r){return(t,e)=>typeof e=="object"?as(r,t,e):((i,s,o)=>{let a=s.hasOwnProperty(o);return s.constructor.createProperty(o,i),a?Object.getOwnPropertyDescriptor(s,o):void 0})(r,t,e)}function B(r){return n({...r,state:!0,attribute:!1})}function z(r,t,e){r.setFormValue(t,t)}function ie(r,t,e="",i){r.setValidity(t,e,i)}function oe(r){r.setValidity({})}function ae(r,t,e="Please fill out this field."){return r?{flags:{customError:!0},message:r}:t?{flags:{valueMissing:!0},message:e}:{flags:{},message:""}}function b(r,t){customElements.get(r)||customElements.define(r,t)}var u=m`
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

  @media (forced-colors: active) {
    .control,
    button,
    a {
      border: 1px solid ButtonText;
    }

    .control:focus-visible,
    button:focus-visible,
    a:focus-visible,
    select:focus-visible,
    textarea:focus-visible,
    input:focus-visible {
      outline: 2px solid Highlight;
    }
  }
`,X=m`
  /* Block fields fill their containing track by default. */
  :host {
    display: block;
    inline-size: 100%;
  }

  .field {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: var(--mb-field-label-gap, var(--mb-space-2, 0.5rem));
    inline-size: 100%;
  }

  .label {
    display: block;
    margin: 0;
    font-size: var(--mb-font-size-sm);
    font-weight: 600;
    line-height: var(--mb-line-height-tight);
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
    line-height: var(--mb-line-height-tight);
    margin: 0;
  }

  .hint {
    color: var(--mb-color-muted);
  }

  .error {
    color: var(--mb-color-danger);
  }

  .control {
    display: block;
    inline-size: 100%;
    max-inline-size: 100%;
    /* Fixed track height so toolbar siblings (input / select / button) match. */
    block-size: var(--mb-control-height, 2.5rem);
    min-block-size: var(--mb-control-height, 2.5rem);
    min-inline-size: 0;
    padding-block: 0;
    padding-inline: var(--mb-control-padding-inline, var(--mb-space-3, 0.75rem));
    border: 1px solid var(--mb-color-border-strong, #6e857a);
    border-radius: var(--mb-radius-md);
    background: var(--mb-color-surface);
    color: var(--mb-color-fg);
    font: inherit;
    line-height: calc(var(--mb-control-height, 2.5rem) - 2px);
    transition:
      border-color var(--mb-transition),
      background-color var(--mb-transition),
      box-shadow var(--mb-transition);
  }

  .control::placeholder {
    color: var(--mb-color-muted);
    opacity: 1;
  }

  .control:hover:not(:disabled) {
    border-color: var(--mb-color-border-hover);
  }

  .control:focus-visible {
    border-color: var(--mb-color-accent);
  }

  .control:disabled {
    opacity: 1;
    cursor: not-allowed;
    background: var(--mb-color-bg);
    color: var(--mb-color-muted);
    border-color: var(--mb-color-border);
  }

  select.control {
    appearance: none;
    -webkit-appearance: none;
    padding-inline-end: var(
      --mb-control-padding-inline-end-select,
      var(--mb-space-5, 1.5rem)
    );
    background-color: var(--mb-color-surface, #fbfcf9);
    background-image: linear-gradient(
        45deg,
        transparent 50%,
        var(--mb-color-muted, #4a5f55) 50%
      ),
      linear-gradient(135deg, var(--mb-color-muted, #4a5f55) 50%, transparent 50%);
    background-position:
      calc(100% - 1rem) 50%,
      calc(100% - 0.65rem) 50%;
    background-size:
      0.35rem 0.35rem,
      0.35rem 0.35rem;
    background-repeat: no-repeat;
  }

  select.control:disabled {
    background-color: var(--mb-color-bg);
  }

  :host([invalid]) .control,
  :host([invalid]) .control:hover:not(:disabled),
  :host([invalid]) .control:focus-visible {
    border-color: var(--mb-color-danger);
  }

  :host([density='compact']) .field {
    gap: 0;
  }

  :host([density='compact']) .control {
    block-size: var(--mb-control-height-sm, 2rem);
    min-block-size: var(--mb-control-height-sm, 2rem);
    padding-block: 0;
    padding-inline: var(--mb-space-2, 0.5rem);
    font-size: var(--mb-font-size-sm);
    line-height: calc(var(--mb-control-height-sm, 2rem) - 2px);
  }

  :host([density='compact']) select.control {
    padding-inline-end: var(--mb-space-5, 1.5rem);
    background-position:
      calc(100% - 0.85rem) 50%,
      calc(100% - 0.5rem) 50%;
  }

  :host([density='compact']) textarea.control {
    block-size: auto;
    min-block-size: var(--mb-control-height-sm, 2rem);
    padding-block: var(--mb-space-2, 0.5rem);
    line-height: var(--mb-line-height, 1.5);
  }
`;function Z(r,t,e){return r?{labelText:r,hideVisually:t,controlAriaLabel:""}:{labelText:"",hideVisually:!1,controlAriaLabel:e}}var ns=Object.defineProperty,V=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&ns(t,e,s),s},D=class extends p{constructor(){super(...arguments),this.variant="primary",this.size="md",this.type="button",this.disabled=!1,this.loading=!1,this.name="",this.value="",this.href="",this.target="",this.rel="",this.iconOnly=!1,this.#t=this.attachInternals(),this.#e=!1}static{this.formAssociated=!0}static{this.styles=[u,m`
      :host {
        display: inline-block;
      }

      .base {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        gap: var(--mb-space-2);
        inline-size: 100%;
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

      .base[aria-busy='true'] {
        cursor: progress;
        opacity: 1;
      }

      :host([size='sm']) .base {
        min-block-size: var(--mb-control-height-sm, 2rem);
        padding-inline: var(--mb-space-3);
        font-size: var(--mb-font-size-sm);
      }

      :host([size='md']) .base {
        min-block-size: var(--mb-control-height, 2.5rem);
        padding-inline: var(--mb-space-4);
        font-size: var(--mb-font-size-md);
      }

      :host([size='lg']) .base {
        min-block-size: 3rem;
        padding-inline: var(--mb-space-5);
        font-size: var(--mb-font-size-lg);
      }

      :host([icon-only][size='sm']) .base {
        min-inline-size: var(--mb-control-height-sm, 2rem);
        padding-inline: 0;
      }

      :host([icon-only][size='md']) .base,
      :host([icon-only]:not([size])) .base {
        min-inline-size: var(--mb-control-height, 2.5rem);
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
        border-color: var(--mb-color-border-strong);
      }

      :host([variant='ghost']) .base {
        background: transparent;
        color: var(--mb-color-accent);
      }

      :host([variant='danger']) .base {
        background: var(--mb-color-danger);
        color: var(--mb-color-on-danger);
      }

      :host([variant='primary']) .base:hover:not(:disabled):not([aria-disabled='true']) {
        background: var(--mb-color-accent-hover);
      }

      :host([variant='primary']) .base:active:not(:disabled):not([aria-disabled='true']) {
        background: var(--mb-color-accent-active);
      }

      :host([variant='secondary']) .base:hover:not(:disabled):not([aria-disabled='true']) {
        background-color: var(--mb-color-surface);
        background-image: linear-gradient(var(--mb-color-hover), var(--mb-color-hover));
        border-color: var(--mb-color-border-hover);
      }

      :host([variant='secondary']) .base:active:not(:disabled):not([aria-disabled='true']) {
        background-color: var(--mb-color-bg);
        background-image: none;
        border-color: var(--mb-color-border-hover);
      }

      :host([variant='ghost']) .base:hover:not(:disabled):not([aria-disabled='true']),
      :host([variant='ghost']) .base:active:not(:disabled):not([aria-disabled='true']) {
        background: var(--mb-color-accent-soft);
      }

      :host([variant='danger']) .base:hover:not(:disabled):not([aria-disabled='true']) {
        background: var(--mb-color-danger-hover);
      }

      :host([variant='danger']) .base:active:not(:disabled):not([aria-disabled='true']) {
        background: var(--mb-color-danger-active);
      }

      .spinner {
        flex: none;
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
    `]}#t;#e;get#r(){return this.disabled||this.loading}get#s(){return!!this.href}get#i(){return this.getAttribute("aria-label")??""}formDisabledCallback(t){this.#e=t,this.requestUpdate()}#o(t){if(this.#r){t.preventDefault(),t.stopImmediatePropagation();return}this.#s||queueMicrotask(()=>{if(t.defaultPrevented||this.#r||!this.isConnected)return;let e=this.#t.form;e&&(this.type==="submit"?(this.name&&z(this.#t,this.value),e.requestSubmit(),queueMicrotask(()=>z(this.#t,null))):this.type==="reset"&&e.reset())})}render(){let t=l`
      ${this.loading?l`<span class="spinner" aria-hidden="true"></span>`:c}
      <slot></slot>
    `,e=this.#i||c;return this.#s?l`
        <a
          part="base"
          class="base"
          href=${this.#r?c:this.href}
          target=${this.target||c}
          rel=${this.rel||(this.target==="_blank"?"noopener noreferrer":c)}
          aria-disabled=${this.#r?"true":"false"}
          aria-busy=${this.loading?"true":"false"}
          aria-label=${e}
          @click=${this.#o}
        >
          ${t}
        </a>
      `:l`
      <button
        part="base"
        class="base"
        type="button"
        ?disabled=${this.#r}
        aria-busy=${this.loading?"true":"false"}
        aria-label=${e}
        @click=${this.#o}
      >
        ${t}
      </button>
    `}};V([n({reflect:!0})],D.prototype,"variant");V([n({reflect:!0})],D.prototype,"size");V([n({reflect:!0})],D.prototype,"type");V([n({type:Boolean,reflect:!0})],D.prototype,"disabled");V([n({type:Boolean,reflect:!0})],D.prototype,"loading");V([n({reflect:!0})],D.prototype,"name");V([n()],D.prototype,"value");V([n({reflect:!0})],D.prototype,"href");V([n({reflect:!0})],D.prototype,"target");V([n({reflect:!0})],D.prototype,"rel");V([n({type:Boolean,reflect:!0,attribute:"icon-only"})],D.prototype,"iconOnly");b("mb-button",D);var ls=Object.defineProperty,ne=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&ls(t,e,s),s},tt=class extends p{constructor(){super(...arguments),this.variant="neutral",this.href="",this.size="sm"}static{this.styles=[u,m`
      :host {
        display: inline-flex;
        max-inline-size: 100%;
      }

      .chip {
        display: inline-flex;
        align-items: center;
        gap: var(--mb-space-1);
        max-inline-size: 100%;
        min-block-size: 1.5rem;
        padding-block: 0.15rem;
        padding-inline: var(--mb-space-2);
        border: 1px solid var(--mb-color-border-strong);
        border-radius: var(--mb-radius-sm);
        background: var(--mb-color-bg);
        color: var(--mb-color-fg);
        font-size: var(--mb-font-size-sm);
        font-weight: 600;
        line-height: var(--mb-line-height-tight);
        text-decoration: none;
        overflow-wrap: anywhere;
        transition:
          background-color var(--mb-transition),
          border-color var(--mb-transition);
      }

      :host([size='md']) .chip {
        min-block-size: var(--mb-control-height-sm, 2rem);
        padding-block: var(--mb-space-1);
        padding-inline: var(--mb-space-3);
        border-radius: var(--mb-radius-md);
      }

      :host([variant='success']) .chip {
        background: var(--mb-color-success-soft);
        color: var(--mb-color-success);
        border-color: var(--mb-color-success);
      }

      :host([variant='warning']) .chip {
        background: var(--mb-color-warning-soft);
        color: var(--mb-color-warning);
        border-color: var(--mb-color-warning);
      }

      :host([variant='danger']) .chip {
        background: var(--mb-color-danger-soft);
        color: var(--mb-color-danger);
        border-color: var(--mb-color-danger);
      }

      :host([variant='info']) .chip {
        background: var(--mb-color-info-soft);
        color: var(--mb-color-info);
        border-color: var(--mb-color-info);
      }

      a.chip:hover {
        background-image: linear-gradient(var(--mb-color-hover), var(--mb-color-hover));
        border-color: var(--mb-color-border-hover);
      }

      a.chip:active {
        background-image: none;
        background-color: var(--mb-color-bg);
        border-color: var(--mb-color-border-hover);
      }

      :host([variant='success']) a.chip:hover,
      :host([variant='warning']) a.chip:hover,
      :host([variant='danger']) a.chip:hover,
      :host([variant='info']) a.chip:hover {
        filter: brightness(0.97);
      }
    `]}render(){return this.href?l`
        <a part="base" class="chip" href=${this.href}>
          <slot></slot>
        </a>
      `:l`<span part="base" class="chip"><slot></slot></span>`}};ne([n({reflect:!0})],tt.prototype,"variant");ne([n({reflect:!0})],tt.prototype,"href");ne([n({reflect:!0})],tt.prototype,"size");b("mb-badge",tt);var cs=Object.defineProperty,hs=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&cs(t,e,s),s},Ut=class extends p{constructor(){super(...arguments),this.variant="info"}static{this.styles=[u,m`
      :host {
        display: block;
        inline-size: 100%;
      }

      .alert {
        padding-block: var(--mb-space-3);
        padding-inline: var(--mb-space-4);
        border: 1px solid var(--mb-color-info);
        border-inline-start-width: 4px;
        border-radius: var(--mb-radius-md);
        background: var(--mb-color-info-soft);
        color: var(--mb-color-fg);
        overflow-wrap: anywhere;
        max-inline-size: 100%;
      }

      :host([variant='success']) .alert {
        background: var(--mb-color-success-soft);
        border-color: var(--mb-color-success);
      }

      :host([variant='warning']) .alert {
        background: var(--mb-color-warning-soft);
        border-color: var(--mb-color-warning);
      }

      :host([variant='danger']) .alert {
        background: var(--mb-color-danger-soft);
        border-color: var(--mb-color-danger);
      }
    `]}get#t(){return this.variant==="warning"||this.variant==="danger"?"alert":"status"}render(){return l`
      <div part="base" class="alert" role=${this.#t}>
        <slot></slot>
      </div>
    `}};hs([n({reflect:!0})],Ut.prototype,"variant");b("mb-alert",Ut);var ds=Object.defineProperty,Le=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&ds(t,e,s),s},At=class extends p{constructor(){super(...arguments),this._hasHeader=!1,this._hasFooter=!1}static{this.styles=[u,m`
      :host {
        display: block;
        inline-size: 100%;
      }

      .card {
        background: var(--mb-color-surface);
        border: 1px solid var(--mb-color-border-strong);
        border-radius: var(--mb-radius-lg);
        box-shadow: var(--mb-shadow-sm);
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
        background: var(--mb-color-bg);
        font-family: var(--mb-font-display);
        font-weight: 600;
        letter-spacing: -0.015em;
        line-height: var(--mb-line-height-tight);
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
    `}};Le([B()],At.prototype,"_hasHeader");Le([B()],At.prototype,"_hasFooter");b("mb-card",At);var P=class{constructor(t){this.host=t,this.formDisabled=!1,this.touched=!1,this.#t=!1,this.internals=t.attachInternals(),t.addController(this)}#t;hostConnected(){}get isDisabled(){return this.host.disabled}captureDefault(t){this.#t||(this.defaultValue=t,this.#t=!0)}formDisabledCallback(t){this.formDisabled=t,this.host.requestUpdate()}markTouched(){this.touched=!0}resetInteraction(){this.touched=!1}applyConstraintValidity(t,e,i,s,o){let a=ae(t,e,s),d=t?a.flags:o?.flags?o.flags:a.flags,h=t?a.message:o?.message?o.message:a.message;return h?(ie(this.internals,d,h,i),!!t||this.touched):(oe(this.internals),!1)}applyNativeOrConstraintValidity(t,e,i,s,o="Please enter a valid value."){let a=i?.validity,d=a&&!a.valid?{badInput:a.badInput,patternMismatch:a.patternMismatch,rangeOverflow:a.rangeOverflow,rangeUnderflow:a.rangeUnderflow,stepMismatch:a.stepMismatch,tooLong:a.tooLong,tooShort:a.tooShort,typeMismatch:a.typeMismatch,valueMissing:a.valueMissing}:{},h=ae(t,e,s),v=t||e?h.flags:d,A=t||e?h.message:a&&!a.valid?i?.validationMessage||o:"";return A?(ie(this.internals,v,A,i),!!t||this.touched):(oe(this.internals),!1)}checkValidity(){return this.internals.checkValidity()}reportValidity(){return this.internals.reportValidity()}};var ps=Object.defineProperty,k=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&ps(t,e,s),s},y=class extends p{constructor(){super(...arguments),this.label="",this.hint="",this.error="",this.value="",this.name="",this.placeholder="",this.type="text",this.disabled=!1,this.required=!1,this.invalid=!1,this.density="default",this.hideLabel=!1,this.min="",this.max="",this.step="",this.accept="",this.multiple=!1,this.pattern="",this.maxLength=null,this.minLength=null,this.autocomplete="",this.readonly=!1,this.missingMessage="Please fill out this field.",this.invalidMessage="Please enter a valid value.",this.#t=new P(this)}static{this.formAssociated=!0}static{this.styles=[u,X,m`
      input[type='file'].control {
        /* File controls need a little block padding for the UA button chrome. */
        padding-block: var(--mb-space-1, 0.25rem);
        line-height: 1.2;
      }
    `]}#t;#e;get#r(){return this.#t.isDisabled}get#s(){return this.type==="file"}get#i(){return this.getAttribute("aria-label")??""}checkValidity(){return this.#t.checkValidity()}reportValidity(){return this.#t.reportValidity()}connectedCallback(){super.connectedCallback(),this.#t.captureDefault(this.value)}firstUpdated(){this.#e=this.renderRoot.querySelector("input")??void 0,this.#a()}updated(t){(t.has("value")||t.has("required")||t.has("error")||t.has("disabled")||t.has("name")||t.has("type")||t.has("min")||t.has("max")||t.has("step")||t.has("multiple")||t.has("pattern")||t.has("maxLength")||t.has("minLength")||t.has("readonly")||t.has("missingMessage")||t.has("invalidMessage"))&&this.#a()}formDisabledCallback(t){this.#t.formDisabledCallback(t)}formResetCallback(){this.#t.resetInteraction(),this.value=this.#t.defaultValue,this.error="",this.invalid=!1,this.#s&&this.#e&&(this.#e.value="")}formStateRestoreCallback(t,e){this.#s||typeof t=="string"&&(this.value=t)}#o(){let t=this.#e?.files;if(!this.name||!t?.length){z(this.#t.internals,null);return}if(t.length===1){z(this.#t.internals,t[0]);return}let e=new FormData;for(let i of t)e.append(this.name,i);z(this.#t.internals,e)}#a(){this.#s?this.#o():z(this.#t.internals,this.name?this.value:null);let t=this.required&&(this.#s?!this.#e?.files?.length:!this.value);this.invalid=this.#t.applyNativeOrConstraintValidity(this.error,t,this.#e,this.missingMessage,this.invalidMessage)}#l(t){let e=t.target;this.#t.markTouched(),this.#s||(this.value=e.value),this.#a(),this.dispatchEvent(new CustomEvent("mb-input",{detail:{value:this.value,files:e.files},bubbles:!0,composed:!0}))}#n(t){let e=t.target;this.#t.markTouched(),this.#s||(this.value=e.value),this.#a(),this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value,files:e.files},bubbles:!0,composed:!0}))}#d(t){if(t.key!=="Enter"||t.defaultPrevented||this.#s)return;let e=this.#t.internals.form;e&&(t.preventDefault(),e.requestSubmit())}render(){let t=[this.hint&&!this.error?"hint":"",this.error?"error":""].filter(Boolean).join(" "),{labelText:e,hideVisually:i,controlAriaLabel:s}=Z(this.label,this.hideLabel,this.#i);return l`
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
          .value=${this.#s?"":this.value}
          name=${this.name||c}
          placeholder=${this.placeholder||c}
          min=${this.type==="number"&&this.min!==""?this.min:c}
          max=${this.type==="number"&&this.max!==""?this.max:c}
          step=${this.type==="number"&&this.step!==""?this.step:c}
          accept=${this.#s&&this.accept?this.accept:c}
          pattern=${!this.#s&&this.pattern?this.pattern:c}
          maxlength=${!this.#s&&this.maxLength!=null?this.maxLength:c}
          minlength=${!this.#s&&this.minLength!=null?this.minLength:c}
          autocomplete=${!this.#s&&this.autocomplete?this.autocomplete:c}
          ?multiple=${this.#s&&this.multiple}
          ?readonly=${!this.#s&&this.readonly}
          ?disabled=${this.#r}
          ?required=${this.required}
          aria-invalid=${this.invalid?"true":"false"}
          aria-label=${s||c}
          aria-describedby=${t||c}
          @input=${this.#l}
          @change=${this.#n}
          @keydown=${this.#d}
        />
        ${this.hint&&!this.error?l`<p id="hint" class="hint">${this.hint}</p>`:c}
        ${this.error?l`<p id="error" class="error" role="alert">${this.error}</p>`:c}
      </div>
    `}};k([n()],y.prototype,"label");k([n()],y.prototype,"hint");k([n()],y.prototype,"error");k([n()],y.prototype,"value");k([n({reflect:!0})],y.prototype,"name");k([n()],y.prototype,"placeholder");k([n({reflect:!0})],y.prototype,"type");k([n({type:Boolean,reflect:!0})],y.prototype,"disabled");k([n({type:Boolean,reflect:!0})],y.prototype,"required");k([n({type:Boolean,reflect:!0})],y.prototype,"invalid");k([n({reflect:!0})],y.prototype,"density");k([n({type:Boolean,reflect:!0,attribute:"hide-label"})],y.prototype,"hideLabel");k([n()],y.prototype,"min");k([n()],y.prototype,"max");k([n()],y.prototype,"step");k([n()],y.prototype,"accept");k([n({type:Boolean})],y.prototype,"multiple");k([n()],y.prototype,"pattern");k([n({type:Number,attribute:"maxlength"})],y.prototype,"maxLength");k([n({type:Number,attribute:"minlength"})],y.prototype,"minLength");k([n()],y.prototype,"autocomplete");k([n({type:Boolean,reflect:!0})],y.prototype,"readonly");k([n({attribute:"missing-message"})],y.prototype,"missingMessage");k([n({attribute:"invalid-message"})],y.prototype,"invalidMessage");b("mb-input",y);var ms=Object.defineProperty,S=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&ms(t,e,s),s},_=class extends p{constructor(){super(...arguments),this.label="",this.hint="",this.error="",this.value="",this.name="",this.placeholder="",this.disabled=!1,this.required=!1,this.invalid=!1,this.rows=4,this.density="default",this.hideLabel=!1,this.maxLength=null,this.minLength=null,this.autocomplete="",this.readonly=!1,this.missingMessage="Please fill out this field.",this.invalidMessage="Please enter a valid value.",this.#t=new P(this)}static{this.formAssociated=!0}static{this.styles=[u,X,m`
      textarea.control {
        block-size: auto;
        min-block-size: 6rem;
        padding-block: var(--mb-space-2, 0.5rem);
        line-height: var(--mb-line-height, 1.5);
        /* Full-width block; not user-resizable unless hosts opt in later. */
        resize: none;
        field-sizing: fixed;
      }
    `]}#t;#e;get#r(){return this.#t.isDisabled}get#s(){return this.getAttribute("aria-label")??""}checkValidity(){return this.#t.checkValidity()}reportValidity(){return this.#t.reportValidity()}connectedCallback(){super.connectedCallback(),this.#t.captureDefault(this.value)}firstUpdated(){this.#e=this.renderRoot.querySelector("textarea")??void 0,this.#i()}updated(t){(t.has("value")||t.has("required")||t.has("error")||t.has("disabled")||t.has("name")||t.has("maxLength")||t.has("minLength")||t.has("readonly")||t.has("missingMessage")||t.has("invalidMessage"))&&this.#i()}formDisabledCallback(t){this.#t.formDisabledCallback(t)}formResetCallback(){this.#t.resetInteraction(),this.value=this.#t.defaultValue,this.error="",this.invalid=!1,this.#i()}formStateRestoreCallback(t,e){typeof t=="string"&&(this.value=t)}#i(){z(this.#t.internals,this.name?this.value:null);let t=this.required&&!this.value;this.invalid=this.#t.applyNativeOrConstraintValidity(this.error,t,this.#e,this.missingMessage,this.invalidMessage)}#o(t){let e=t.target;this.#t.markTouched(),this.value=e.value,this.#i(),this.dispatchEvent(new CustomEvent("mb-input",{detail:{value:this.value},bubbles:!0,composed:!0}))}#a(t){let e=t.target;this.#t.markTouched(),this.value=e.value,this.#i(),this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value},bubbles:!0,composed:!0}))}render(){let t=[this.hint&&!this.error?"hint":"",this.error?"error":""].filter(Boolean).join(" "),{labelText:e,hideVisually:i,controlAriaLabel:s}=Z(this.label,this.hideLabel,this.#s);return l`
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
          maxlength=${this.maxLength!=null?this.maxLength:c}
          minlength=${this.minLength!=null?this.minLength:c}
          autocomplete=${this.autocomplete||c}
          ?readonly=${this.readonly}
          ?disabled=${this.#r}
          ?required=${this.required}
          aria-invalid=${this.invalid?"true":"false"}
          aria-label=${s||c}
          aria-describedby=${t||c}
          @input=${this.#o}
          @change=${this.#a}
        ></textarea>
        ${this.hint&&!this.error?l`<p id="hint" class="hint">${this.hint}</p>`:c}
        ${this.error?l`<p id="error" class="error" role="alert">${this.error}</p>`:c}
      </div>
    `}};S([n()],_.prototype,"label");S([n()],_.prototype,"hint");S([n()],_.prototype,"error");S([n()],_.prototype,"value");S([n({reflect:!0})],_.prototype,"name");S([n()],_.prototype,"placeholder");S([n({type:Boolean,reflect:!0})],_.prototype,"disabled");S([n({type:Boolean,reflect:!0})],_.prototype,"required");S([n({type:Boolean,reflect:!0})],_.prototype,"invalid");S([n({type:Number})],_.prototype,"rows");S([n({reflect:!0})],_.prototype,"density");S([n({type:Boolean,reflect:!0,attribute:"hide-label"})],_.prototype,"hideLabel");S([n({type:Number,attribute:"maxlength"})],_.prototype,"maxLength");S([n({type:Number,attribute:"minlength"})],_.prototype,"minLength");S([n()],_.prototype,"autocomplete");S([n({type:Boolean,reflect:!0})],_.prototype,"readonly");S([n({attribute:"missing-message"})],_.prototype,"missingMessage");S([n({attribute:"invalid-message"})],_.prototype,"invalidMessage");b("mb-textarea",_);var bs=Object.defineProperty,j=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&bs(t,e,s),s},R=class extends p{constructor(){super(...arguments),this.label="",this.error="",this.name="",this.value="on",this.checked=!1,this.indeterminate=!1,this.disabled=!1,this.required=!1,this.invalid=!1,this.missingMessage="Please check this box.",this.#t=new P(this)}static{this.formAssociated=!0}static{this.styles=[u,m`
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
        flex: none;
        margin-block-start: 0;
        accent-color: var(--mb-color-accent);
        inline-size: 1.5rem;
        block-size: 1.5rem;
      }

      input:disabled {
        cursor: not-allowed;
      }

      :host([invalid]) input {
        outline: 2px solid var(--mb-color-danger);
        outline-offset: 2px;
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
    `]}#t;#e;get#r(){return this.#t.isDisabled}checkValidity(){return this.#t.checkValidity()}reportValidity(){return this.#t.reportValidity()}connectedCallback(){super.connectedCallback(),this.#t.captureDefault(this.checked)}firstUpdated(){this.#e=this.renderRoot.querySelector("input")??void 0,this.#s(),this.#i()}updated(t){t.has("indeterminate")&&this.#s(),(t.has("checked")||t.has("value")||t.has("required")||t.has("error")||t.has("disabled")||t.has("name")||t.has("missingMessage"))&&this.#i()}formDisabledCallback(t){this.#t.formDisabledCallback(t)}formResetCallback(){this.#t.resetInteraction(),this.checked=this.#t.defaultValue,this.indeterminate=!1,this.error="",this.invalid=!1,this.#i()}formStateRestoreCallback(t,e){if(t==null){this.checked=!1;return}typeof t=="string"&&(this.checked=!0,t&&(this.value=t))}#s(){this.#e&&(this.#e.indeterminate=this.indeterminate)}#i(){z(this.#t.internals,this.name&&this.checked?this.value:null);let t=this.required&&!this.checked;this.invalid=this.#t.applyConstraintValidity(this.error,t,this.#e,this.missingMessage)}#o(t){let e=t.target;this.#t.markTouched(),this.checked=e.checked,this.indeterminate=!1,this.#i(),this.dispatchEvent(new CustomEvent("mb-change",{detail:{checked:this.checked,value:this.value},bubbles:!0,composed:!0}))}render(){let t=this.error?"error":"";return l`
      <label part="label">
        <input
          part="control"
          type="checkbox"
          .checked=${this.checked}
          name=${this.name||c}
          value=${this.value}
          ?disabled=${this.#r}
          ?required=${this.required}
          aria-invalid=${this.invalid?"true":"false"}
          aria-describedby=${t||c}
          @change=${this.#o}
        />
        <span>${this.label}<slot></slot></span>
      </label>
      ${this.error?l`<p id="error" class="error" role="alert">${this.error}</p>`:c}
    `}};j([n()],R.prototype,"label");j([n()],R.prototype,"error");j([n({reflect:!0})],R.prototype,"name");j([n()],R.prototype,"value");j([n({type:Boolean,reflect:!0})],R.prototype,"checked");j([n({type:Boolean,reflect:!0})],R.prototype,"indeterminate");j([n({type:Boolean,reflect:!0})],R.prototype,"disabled");j([n({type:Boolean,reflect:!0})],R.prototype,"required");j([n({type:Boolean,reflect:!0})],R.prototype,"invalid");j([n({attribute:"missing-message"})],R.prototype,"missingMessage");b("mb-checkbox",R);var De={ATTRIBUTE:1,CHILD:2,PROPERTY:3,BOOLEAN_ATTRIBUTE:4,EVENT:5,ELEMENT:6},Pe=r=>(...t)=>({_$litDirective$:r,values:t}),It=class{constructor(t){}get _$AU(){return this._$AM._$AU}_$AT(t,e,i){this._$Ct=t,this._$AM=e,this._$Ci=i}_$AS(t,e){return this.update(t,e)}update(t,e){return this.render(...e)}};var{I:us}=Ee,Oe=r=>r;var Re=()=>document.createComment(""),bt=(r,t,e)=>{let i=r._$AA.parentNode,s=t===void 0?r._$AB:t._$AA;if(e===void 0){let o=i.insertBefore(Re(),s),a=i.insertBefore(Re(),s);e=new us(o,a,r,r.options)}else{let o=e._$AB.nextSibling,a=e._$AM,d=a!==r;if(d){let h;e._$AQ?.(r),e._$AM=r,e._$AP!==void 0&&(h=r._$AU)!==a._$AU&&e._$AP(h)}if(o!==s||d){let h=e._$AA;for(;h!==o;){let v=Oe(h).nextSibling;Oe(i).insertBefore(h,s),h=v}}}return e},et=(r,t,e=r)=>(r._$AI(t,e),r),fs={},qe=(r,t=fs)=>r._$AH=t,Te=r=>r._$AH,Ht=r=>{r._$AR(),r._$AA.remove()};var Be=(r,t,e)=>{let i=new Map;for(let s=t;s<=e;s++)i.set(r[s],s);return i},Y=Pe(class extends It{constructor(r){if(super(r),r.type!==De.CHILD)throw Error("repeat() can only be used in text expressions")}dt(r,t,e){let i;e===void 0?e=t:t!==void 0&&(i=t);let s=[],o=[],a=0;for(let d of r)s[a]=i?i(d,a):a,o[a]=e(d,a),a++;return{values:o,keys:s}}render(r,t,e){return this.dt(r,t,e).values}update(r,[t,e,i]){let s=Te(r),{values:o,keys:a}=this.dt(t,e,i);if(!Array.isArray(s))return this.ut=a,o;let d=this.ut??=[],h=[],v,A,f=0,x=s.length-1,g=0,C=o.length-1;for(;f<=x&&g<=C;)if(s[f]===null)f++;else if(s[x]===null)x--;else if(d[f]===a[g])h[g]=et(s[f],o[g]),f++,g++;else if(d[x]===a[C])h[C]=et(s[x],o[C]),x--,C--;else if(d[f]===a[C])h[C]=et(s[f],o[C]),bt(r,h[C+1],s[f]),f++,C--;else if(d[x]===a[g])h[g]=et(s[x],o[g]),bt(r,s[f],s[x]),x--,g++;else if(v===void 0&&(v=Be(a,g,C),A=Be(d,f,x)),v.has(d[f]))if(v.has(d[x])){let K=A.get(a[g]),Wt=K!==void 0?s[K]:null;if(Wt===null){let pe=bt(r,s[f]);et(pe,o[g]),h[g]=pe}else h[g]=et(Wt,o[g]),bt(r,s[f],Wt),s[K]=null;g++}else Ht(s[x]),x--;else Ht(s[f]),f++;for(;g<=C;){let K=bt(r,h[C+1]);et(K,o[g]),h[g++]=K}for(;f<=x;){let K=s[f++];K!==null&&Ht(K)}return this.ut=a,qe(r,h),Q}});function U(r,t,e){if(!r)return[];try{let i=JSON.parse(r);return Array.isArray(i)?i.filter(t).map(e):[]}catch{return[]}}function I(r){return{fromAttribute:r,toAttribute(t){return t?.length?JSON.stringify(t):null}}}var vs=Object.defineProperty,L=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&vs(t,e,s),s};function gs(r){return U(r,t=>!!t&&typeof t=="object"&&typeof t.value=="string"&&typeof t.label=="string",t=>({value:t.value,label:t.label,disabled:!!t.disabled}))}var E=class extends p{constructor(){super(...arguments),this.label="",this.hint="",this.error="",this.value="",this.name="",this.disabled=!1,this.required=!1,this.invalid=!1,this.density="default",this.hideLabel=!1,this.placeholder="",this.missingMessage="Please select an option.",this.invalidMessage="Please select a valid option.",this.options=[],this._slottedOptions=[],this.#t=new P(this)}static{this.formAssociated=!0}static{this.styles=[u,X,m`
      /* Options are mirrored into the shadow <select>; keep light-DOM slots invisible. */
      slot {
        display: none;
      }
    `]}#t;#e;get#r(){return this.#t.isDisabled}get#s(){return this._slottedOptions.length?this._slottedOptions:this.options}get#i(){return this.#s.filter(t=>t.value!=="")}get#o(){let t=this.#s.find(e=>e.value==="");return t?.label?t.label:this.placeholder}get#a(){return this.getAttribute("aria-label")??""}checkValidity(){return this.#t.checkValidity()}reportValidity(){return this.#t.reportValidity()}connectedCallback(){super.connectedCallback(),this.#t.captureDefault(this.value),this.#n()}firstUpdated(){this.#e=this.renderRoot.querySelector("select")??void 0,this.#c()}updated(t){(t.has("value")||t.has("required")||t.has("error")||t.has("options")||t.has("_slottedOptions")||t.has("disabled")||t.has("name")||t.has("missingMessage")||t.has("invalidMessage"))&&this.#c()}formDisabledCallback(t){this.#t.formDisabledCallback(t)}formResetCallback(){this.#t.resetInteraction(),this.value=this.#t.defaultValue,this.error="",this.invalid=!1,this.#c()}formStateRestoreCallback(t,e){typeof t=="string"&&(this.value=t)}#l(t){return t instanceof HTMLOptionElement?{value:t.value,label:t.label||t.textContent?.trim()||t.value,disabled:t.disabled}:null}#n(){let t=[...this.querySelectorAll(":scope > option")].map(e=>this.#l(e)).filter(e=>e!=null);t.length&&(this._slottedOptions=t)}#d(){let t=this.renderRoot.querySelector('slot[name="options"]'),e=this.renderRoot.querySelector("slot:not([name])"),i=[...t?.assignedElements({flatten:!0})??[],...e?.assignedElements({flatten:!0})??[]].map(a=>this.#l(a)).filter(a=>a!=null),s=JSON.stringify(this._slottedOptions),o=JSON.stringify(i);s!==o&&(this._slottedOptions=i)}#p(){this.#d()}#c(){this.#e&&this.#e.value!==this.value&&(this.#e.value=this.value);let t=!this.value||this.#s.some(s=>s.value===this.value&&!s.disabled);z(this.#t.internals,this.name&&t?this.value:null);let e=this.required&&!this.value,i=!!this.value&&!t;this.invalid=this.#t.applyConstraintValidity(this.error,e,this.#e,this.missingMessage,i?{flags:{badInput:!0},message:this.invalidMessage}:null)}#h(t){let e=t.target;this.#t.markTouched(),this.value=e.value,this.#c(),this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value},bubbles:!0,composed:!0}))}render(){let t=[this.hint&&!this.error?"hint":"",this.error?"error":""].filter(Boolean).join(" "),{labelText:e,hideVisually:i,controlAriaLabel:s}=Z(this.label,this.hideLabel,this.#a);return l`
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
          ?disabled=${this.#r}
          ?required=${this.required}
          aria-invalid=${this.invalid?"true":"false"}
          aria-label=${s||c}
          aria-describedby=${t||c}
          .value=${this.value}
          @change=${this.#h}
        >
          <option value="" ?disabled=${this.required}>${this.#o}</option>
          ${Y(this.#i,o=>o.value,o=>l`
              <option value=${o.value} ?disabled=${!!o.disabled}>
                ${o.label}
              </option>
            `)}
        </select>
        ${this.hint&&!this.error?l`<p id="hint" class="hint">${this.hint}</p>`:c}
        ${this.error?l`<p id="error" class="error" role="alert">${this.error}</p>`:c}
      </div>
      <slot name="options" @slotchange=${this.#p}></slot>
      <slot @slotchange=${this.#p}></slot>
    `}};L([n()],E.prototype,"label");L([n()],E.prototype,"hint");L([n()],E.prototype,"error");L([n()],E.prototype,"value");L([n({reflect:!0})],E.prototype,"name");L([n({type:Boolean,reflect:!0})],E.prototype,"disabled");L([n({type:Boolean,reflect:!0})],E.prototype,"required");L([n({type:Boolean,reflect:!0})],E.prototype,"invalid");L([n({reflect:!0})],E.prototype,"density");L([n({type:Boolean,reflect:!0,attribute:"hide-label"})],E.prototype,"hideLabel");L([n()],E.prototype,"placeholder");L([n({attribute:"missing-message"})],E.prototype,"missingMessage");L([n({attribute:"invalid-message"})],E.prototype,"invalidMessage");L([n({attribute:"options",converter:I(gs)})],E.prototype,"options");L([B()],E.prototype,"_slottedOptions");b("mb-select",E);var ys=Object.defineProperty,w=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&ys(t,e,s),s};function $s(r){return U(r,t=>!!t&&typeof t=="object"&&typeof t.value=="string"&&typeof t.label=="string",t=>({value:t.value,label:t.label,disabled:!!t.disabled,group:typeof t.group=="string"?t.group:void 0,href:typeof t.href=="string"?t.href:void 0}))}var $=class extends p{constructor(){super(...arguments),this.label="",this.hint="",this.error="",this.value="",this.name="",this.placeholder="",this.type="search",this.disabled=!1,this.required=!1,this.invalid=!1,this.density="default",this.hideLabel=!1,this.open=!1,this.loading=!1,this.options=[],this.emptyMessage="No results",this.loadingMessage="Loading\u2026",this.closeOnSelect=!0,this.closeOnBlur=!0,this.missingMessage="Please fill out this field.",this.invalidMessage="Please enter a valid value.",this._slottedOptions=[],this._activeIndex=-1,this.#t=new P(this),this.#r=`mb-combobox-list-${Math.random().toString(36).slice(2,9)}`,this.#s=0,this.#i=!1,this.#C=t=>{!this.open||!this.closeOnBlur||t.composedPath().includes(this)||this.#y()}}static{this.formAssociated=!0}static{this.styles=[u,X,m`
      :host {
        position: relative;
      }

      .wrap {
        position: relative;
        inline-size: 100%;
      }

      .panel {
        position: absolute;
        inset-inline: 0;
        top: calc(100% + var(--mb-space-1, 0.25rem));
        z-index: 40;
        margin: 0;
        padding: var(--mb-space-1, 0.25rem);
        list-style: none;
        border: 1px solid var(--mb-color-border-strong);
        border-radius: var(--mb-radius-md);
        background: var(--mb-color-surface);
        color: var(--mb-color-fg);
        box-shadow: var(--mb-shadow);
        max-block-size: min(18rem, 50vh);
        overflow: auto;
      }

      .group {
        padding-block: var(--mb-space-2, 0.5rem) var(--mb-space-1, 0.25rem);
        padding-inline: var(--mb-space-3, 0.75rem);
        font-size: var(--mb-font-size-sm);
        font-weight: 600;
        color: var(--mb-color-muted);
        line-height: var(--mb-line-height-tight);
      }

      .option {
        display: block;
        inline-size: 100%;
        margin: 0;
        padding-block: var(--mb-space-2, 0.5rem);
        padding-inline: var(--mb-space-3, 0.75rem);
        border: 0;
        border-radius: var(--mb-radius-sm);
        background: transparent;
        color: inherit;
        font: inherit;
        text-align: start;
        cursor: pointer;
        text-decoration: none;
        line-height: var(--mb-line-height-tight);
      }

      .option:hover:not([aria-disabled='true']),
      .option[data-active]:not([aria-disabled='true']) {
        background: var(--mb-color-hover);
      }

      .option[aria-disabled='true'] {
        color: var(--mb-color-muted);
        cursor: not-allowed;
      }

      .status {
        display: flex;
        align-items: center;
        gap: var(--mb-space-2, 0.5rem);
        padding-block: var(--mb-space-3, 0.75rem);
        padding-inline: var(--mb-space-3, 0.75rem);
        color: var(--mb-color-muted);
        font-size: var(--mb-font-size-sm);
        line-height: var(--mb-line-height-tight);
      }

      .spinner {
        flex-shrink: 0;
        inline-size: 0.9rem;
        block-size: 0.9rem;
        border: 2px solid var(--mb-color-border);
        border-inline-end-color: var(--mb-color-accent);
        border-radius: 50%;
        animation: spin 0.7s linear infinite;
      }

      @keyframes spin {
        to {
          transform: rotate(360deg);
        }
      }

      /* Options are mirrored into the shadow listbox; keep light-DOM slots invisible. */
      slot {
        display: none;
      }
    `]}#t;#e;#r;#s;#i;get#o(){return this.#t.isDisabled}get#a(){return this.getAttribute("aria-label")??""}get#l(){return this._slottedOptions.length?this._slottedOptions:this.options}get#n(){return this.#l.map((t,e)=>({...t,index:e,id:`${this.#r}-opt-${e}`}))}get#d(){return this.#n.filter(t=>!t.disabled).map(t=>t.index)}get#p(){return this.#n[this._activeIndex]}get#c(){return this.open&&!this.#o}checkValidity(){return this.#t.checkValidity()}reportValidity(){return this.#t.reportValidity()}connectedCallback(){super.connectedCallback(),this.#t.captureDefault(this.value),this.#$(),document.addEventListener("pointerdown",this.#C,!0)}disconnectedCallback(){super.disconnectedCallback(),document.removeEventListener("pointerdown",this.#C,!0),window.clearTimeout(this.#s)}firstUpdated(){this.#e=this.renderRoot.querySelector("input")??void 0,this.#u()}updated(t){(t.has("value")||t.has("required")||t.has("error")||t.has("disabled")||t.has("name")||t.has("missingMessage")||t.has("invalidMessage"))&&this.#u(),(t.has("options")||t.has("_slottedOptions")||t.has("loading"))&&this.open&&this.#x(),t.has("open")&&this.open&&this.#x()}formDisabledCallback(t){this.#t.formDisabledCallback(t)}formResetCallback(){this.#t.resetInteraction(),this.value=this.#t.defaultValue,this.error="",this.invalid=!1,this.open=!1,this._activeIndex=-1,this.#u()}formStateRestoreCallback(t,e){typeof t=="string"&&(this.value=t)}#h(t){if(!(t instanceof HTMLOptionElement))return null;let e=t.getAttribute("data-group")||void 0,i=t.getAttribute("data-href")||void 0;return{value:t.value,label:t.label||t.textContent?.trim()||t.value,disabled:t.disabled,group:e||void 0,href:i||void 0}}#$(){let t=[...this.querySelectorAll(":scope > option")].map(e=>this.#h(e)).filter(e=>e!=null);t.length&&(this._slottedOptions=t)}#S(){let t=this.renderRoot.querySelector('slot[name="options"]'),e=this.renderRoot.querySelector("slot:not([name])"),i=[...t?.assignedElements({flatten:!0})??[],...e?.assignedElements({flatten:!0})??[]].map(a=>this.#h(a)).filter(a=>a!=null),s=JSON.stringify(this._slottedOptions),o=JSON.stringify(i);s!==o&&(this._slottedOptions=i)}#m(){this.#S()}#u(){z(this.#t.internals,this.name?this.value:null);let t=this.required&&!this.value;this.invalid=this.#t.applyNativeOrConstraintValidity(this.error,t,this.#e,this.missingMessage,this.invalidMessage)}#x(){let t=this.#d;if(!t.length){this._activeIndex=-1;return}t.includes(this._activeIndex)||(this._activeIndex=t[0])}#f(t){let e=this.#d;if(!e.length){this._activeIndex=-1;return}let i=e.indexOf(this._activeIndex),s;i===-1?s=t>0?0:e.length-1:s=(i+t+e.length)%e.length,this._activeIndex=e[s],this.#w()}#k(t){let e=this.#d;if(!e.length){this._activeIndex=-1;return}this._activeIndex=t==="start"?e[0]:e[e.length-1],this.#w()}#w(){let t=this.#p;if(!t)return;let e=this.renderRoot.querySelector(`#${CSS.escape(t.id)}`);e instanceof HTMLElement&&e.scrollIntoView({block:"nearest"})}#b(){this.#o||(this.open=!0,this.#x())}#y(){this.open=!1,this._activeIndex=-1}#A(t){t.disabled||(this.#t.markTouched(),this.value=t.label,this.#u(),this.dispatchEvent(new CustomEvent("mb-select",{detail:{value:t.value,label:t.label,href:t.href},bubbles:!0,composed:!0})),this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value},bubbles:!0,composed:!0})),this.closeOnSelect&&this.#y())}#_(t){let e=t.target;this.#t.markTouched(),this.value=e.value,this.#u(),this.#b(),this.dispatchEvent(new CustomEvent("mb-input",{detail:{value:this.value},bubbles:!0,composed:!0}))}#v(t){let e=t.target;this.#t.markTouched(),this.value=e.value,this.#u(),this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value},bubbles:!0,composed:!0}))}#E(){window.clearTimeout(this.#s),(this.loading||this.#l.length)&&this.#b()}#g(){!this.closeOnBlur||this.#i||(window.clearTimeout(this.#s),this.#s=window.setTimeout(()=>{this.matches(":focus-within")||this.#y()},0))}#C;#M(t){if(!this.#o)switch(t.key){case"ArrowDown":{t.preventDefault(),this.open?this.#f(1):this.#b();break}case"ArrowUp":{t.preventDefault(),this.open?this.#f(-1):this.#b();break}case"Home":{this.open&&(t.preventDefault(),this.#k("start"));break}case"End":{this.open&&(t.preventDefault(),this.#k("end"));break}case"Enter":{this.open&&this.#p&&(t.preventDefault(),this.#A(this.#p));break}case"Escape":{this.open&&(t.preventDefault(),this.#y());break}}}#z(t){t.preventDefault(),this.#i=!0}#L(t){this.#i=!1,this.#A(t),this.#e?.focus()}#D(){let t=this.#n,e=[];for(let i of t){let s=e[e.length-1];s&&s.group===i.group?s.items.push(i):e.push({group:i.group,items:[i]})}return Y(e,(i,s)=>`${i.group??""}::${s}`,i=>l`
        ${i.group?l`<div class="group" role="presentation">${i.group}</div>`:c}
        ${Y(i.items,s=>s.id,s=>l`
            <div
              id=${s.id}
              class="option"
              part="option"
              role="option"
              tabindex="-1"
              ?data-active=${s.index===this._activeIndex}
              aria-selected=${s.index===this._activeIndex?"true":"false"}
              aria-disabled=${s.disabled?"true":"false"}
              @pointerdown=${this.#z}
              @click=${()=>this.#L(s)}
            >
              ${s.label}
            </div>
          `)}
      `)}render(){let t=[this.hint&&!this.error?"hint":"",this.error?"error":""].filter(Boolean).join(" "),{labelText:e,hideVisually:i,controlAriaLabel:s}=Z(this.label,this.hideLabel,this.#a),o=this.#p,a=this.#c&&!this.loading&&this.#n.length===0,d=this.#c&&this.#n.length>0;return l`
      <div class="field">
        ${e?l`<label
              part="label"
              class="label${i?" visually-hidden":""}"
              for="control"
              >${e}</label
            >`:c}
        <div class="wrap">
          <input
            id="control"
            part="control"
            class="control"
            role="combobox"
            .type=${this.type}
            .value=${this.value}
            name=${this.name||c}
            placeholder=${this.placeholder||c}
            autocomplete="off"
            ?disabled=${this.#o}
            ?required=${this.required}
            aria-invalid=${this.invalid?"true":"false"}
            aria-label=${s||c}
            aria-describedby=${t||c}
            aria-expanded=${this.#c?"true":"false"}
            aria-controls=${this.#r}
            aria-autocomplete="list"
            aria-activedescendant=${this.#c&&o?o.id:c}
            @input=${this.#_}
            @change=${this.#v}
            @focus=${this.#E}
            @blur=${this.#g}
            @keydown=${this.#M}
          />
          ${this.#c?l`
                <div
                  id=${this.#r}
                  class="panel"
                  part="panel"
                  role="listbox"
                  aria-label=${e||s||"Suggestions"}
                >
                  ${this.loading?l`
                        <div class="status" part="status" role="status" aria-live="polite">
                          <span class="spinner" aria-hidden="true"></span>
                          <span>${this.loadingMessage}</span>
                        </div>
                      `:c}
                  ${d?this.#D():c}
                  ${a?l`
                        <div class="status" part="empty" role="status" aria-live="polite">
                          ${this.emptyMessage}
                        </div>
                      `:c}
                </div>
              `:c}
        </div>
        ${this.hint&&!this.error?l`<p id="hint" class="hint">${this.hint}</p>`:c}
        ${this.error?l`<p id="error" class="error" role="alert">${this.error}</p>`:c}
      </div>
      <slot name="options" @slotchange=${this.#m}></slot>
      <slot @slotchange=${this.#m}></slot>
    `}};w([n()],$.prototype,"label");w([n()],$.prototype,"hint");w([n()],$.prototype,"error");w([n()],$.prototype,"value");w([n({reflect:!0})],$.prototype,"name");w([n()],$.prototype,"placeholder");w([n({reflect:!0})],$.prototype,"type");w([n({type:Boolean,reflect:!0})],$.prototype,"disabled");w([n({type:Boolean,reflect:!0})],$.prototype,"required");w([n({type:Boolean,reflect:!0})],$.prototype,"invalid");w([n({reflect:!0})],$.prototype,"density");w([n({type:Boolean,reflect:!0,attribute:"hide-label"})],$.prototype,"hideLabel");w([n({type:Boolean,reflect:!0})],$.prototype,"open");w([n({type:Boolean,reflect:!0})],$.prototype,"loading");w([n({attribute:"options",converter:I($s)})],$.prototype,"options");w([n({attribute:"empty-message"})],$.prototype,"emptyMessage");w([n({attribute:"loading-message"})],$.prototype,"loadingMessage");w([n({type:Boolean,attribute:"close-on-select"})],$.prototype,"closeOnSelect");w([n({type:Boolean,attribute:"close-on-blur"})],$.prototype,"closeOnBlur");w([n({attribute:"missing-message"})],$.prototype,"missingMessage");w([n({attribute:"invalid-message"})],$.prototype,"invalidMessage");w([B()],$.prototype,"_slottedOptions");w([B()],$.prototype,"_activeIndex");b("mb-combobox",$);var xs=Object.defineProperty,le=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&xs(t,e,s),s},ut=class extends p{constructor(){super(...arguments),this.open=!1,this.heading="",this.closeLabel="Close",this.#e=!1}static{this.styles=[u,m`
      :host {
        display: contents;
      }

      dialog {
        border: 1px solid var(--mb-color-border-strong);
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
        font-weight: 600;
        letter-spacing: -0.02em;
        line-height: var(--mb-line-height-tight);
        margin: 0;
        min-inline-size: 0;
        flex: 1;
        overflow-wrap: anywhere;
      }

      .close {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
        min-inline-size: 2rem;
        min-block-size: 2rem;
        margin: 0;
        padding: 0;
        border: 0;
        border-radius: var(--mb-radius-sm);
        background: transparent;
        color: var(--mb-color-fg);
        font-size: 1.25rem;
        line-height: 1;
        cursor: pointer;
        transition: background-color var(--mb-transition);
      }

      .close:hover {
        background: var(--mb-color-hover);
      }

      .close:active {
        background: var(--mb-color-border);
      }
    `]}#t;#e;firstUpdated(){this.#t=this.renderRoot.querySelector("dialog")??void 0,this.#t?.addEventListener("close",()=>{this.#e||(this.open&&(this.open=!1),this.#s())}),this.#r()}updated(t){t.has("open")&&this.#r()}#r(){let t=this.#t;t&&(this.open&&!t.open?t.showModal():!this.open&&t.open&&(this.#e=!0,t.close(),this.#e=!1,this.#s()))}#s(){this.dispatchEvent(new CustomEvent("mb-close",{bubbles:!0,composed:!0}))}close(){!this.open&&!this.#t?.open||(this.open=!1)}#i(){this.close()}render(){return l`
      <dialog part="dialog" aria-labelledby="title" aria-modal="true">
        <div class="panel">
          <div class="header">
            <h2 class="title" id="title">${this.heading}<slot name="heading"></slot></h2>
            <button
              class="close"
              type="button"
              aria-label=${this.closeLabel}
              @click=${this.#i}
            >
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
    `}};le([n({type:Boolean,reflect:!0})],ut.prototype,"open");le([n()],ut.prototype,"heading");le([n({attribute:"close-label"})],ut.prototype,"closeLabel");b("mb-modal",ut);var ks=Object.defineProperty,_t=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&ks(t,e,s),s},st=class extends p{constructor(){super(...arguments),this.value=0,this.max=100,this.percent=null,this.label="",this.fallbackLabel="Progress"}static{this.styles=[u,m`
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
        block-size: 0.625rem;
        border: 1px solid var(--mb-color-border-strong);
        border-radius: 999px;
        background: var(--mb-color-bg);
        overflow: clip;
      }

      .bar {
        block-size: 100%;
        background: var(--mb-color-accent);
        border-radius: inherit;
        transition: inline-size var(--mb-transition);
      }
    `]}get#t(){if(this.percent!=null&&Number.isFinite(this.percent))return Math.min(100,Math.max(0,this.percent));let t=Number.isFinite(this.value)?this.value:0;return Math.min(100,Math.max(0,t/this.#r*100))}get#e(){if(this.percent!=null&&Number.isFinite(this.percent))return this.#t;let t=Number.isFinite(this.value)?this.value:0;return Math.min(this.#r,Math.max(0,t))}get#r(){return this.percent!=null&&Number.isFinite(this.percent)?100:Number.isFinite(this.max)&&this.max>0?this.max:100}render(){let t=this.#t;return l`
      <div class="wrap">
        ${this.label?l`<div part="label" class="label" id="label">${this.label}</div>`:c}
        <div
          part="track"
          class="track"
          role="progressbar"
          aria-valuemin="0"
          aria-valuenow=${this.#e}
          aria-valuemax=${this.#r}
          aria-labelledby=${this.label?"label":c}
          aria-label=${this.label?c:this.getAttribute("aria-label")||this.fallbackLabel}
        >
          <div part="bar" class="bar" style="inline-size: ${t}%"></div>
        </div>
        <slot></slot>
      </div>
    `}};_t([n({type:Number})],st.prototype,"value");_t([n({type:Number})],st.prototype,"max");_t([n({type:Number})],st.prototype,"percent");_t([n()],st.prototype,"label");_t([n({attribute:"fallback-label"})],st.prototype,"fallbackLabel");b("mb-progress",st);var ws=Object.defineProperty,As=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&ws(t,e,s),s},Ft=class extends p{constructor(){super(...arguments),this.label="Filters"}static{this.styles=[u,m`
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
        align-items: stretch;
        min-inline-size: 100%;
        min-block-size: var(--mb-control-height, 2.5rem);
        gap: 0;
        padding: var(--mb-space-1);
        border: 1px solid var(--mb-color-border-strong);
        border-radius: var(--mb-radius-md);
        background: var(--mb-color-surface);
        box-sizing: border-box;
      }

      ::slotted(a),
      ::slotted(button) {
        appearance: none;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        min-block-size: calc(
          var(--mb-control-height, 2.5rem) - 2 * var(--mb-space-1) - 2px
        );
        border: 0;
        background: transparent;
        color: var(--mb-color-muted);
        font: inherit;
        font-weight: 600;
        font-size: var(--mb-font-size-sm);
        text-decoration: none;
        padding-block: 0;
        padding-inline: var(--mb-space-3);
        border-radius: var(--mb-radius-sm);
        white-space: nowrap;
        cursor: pointer;
      }

      ::slotted(a:hover),
      ::slotted(button:hover) {
        background: var(--mb-color-hover);
        color: var(--mb-color-fg);
      }

      ::slotted(a:focus-visible),
      ::slotted(button:focus-visible) {
        outline: var(--mb-focus-ring);
        outline-offset: var(--mb-focus-offset);
      }

      ::slotted([aria-current='page']),
      ::slotted([aria-selected='true']),
      ::slotted(.is-active),
      ::slotted([aria-current='page']:hover),
      ::slotted([aria-selected='true']:hover),
      ::slotted(.is-active:hover) {
        background: var(--mb-color-accent-soft);
        color: var(--mb-color-accent);
      }
    `]}render(){return l`
      <nav part="nav" class="scroller" aria-label=${this.label}>
        <div part="list" class="list">
          <slot></slot>
        </div>
      </nav>
    `}};As([n()],Ft.prototype,"label");b("mb-segmented-control",Ft);var _s=Object.defineProperty,Cs=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&_s(t,e,s),s},Kt=class extends p{constructor(){super(...arguments),this.heading=""}static{this.styles=[u,m`
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
        border: 1px dashed var(--mb-color-border-strong);
        border-radius: var(--mb-radius-lg);
        background: var(--mb-color-surface);
      }

      .heading {
        margin: 0;
        font-family: var(--mb-font-display);
        font-size: var(--mb-font-size-lg);
        font-weight: 600;
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
    `}};Cs([n()],Kt.prototype,"heading");b("mb-empty-state",Kt);var zs=Object.defineProperty,rt=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&zs(t,e,s),s},N=class extends p{constructor(){super(...arguments),this.prevUrl="",this.nextUrl="",this.prevDisabled=!1,this.nextDisabled=!1,this.status="",this.prevLabel="Previous",this.nextLabel="Next",this.label="Pagination"}static{this.styles=[u,m`
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
        justify-content: center;
        min-block-size: 2.25rem;
        min-inline-size: 2.25rem;
        padding-inline: var(--mb-space-3);
        border: 1px solid var(--mb-color-border-strong);
        border-radius: var(--mb-radius-md);
        background: var(--mb-color-surface);
        color: var(--mb-color-fg);
        font-size: var(--mb-font-size-sm);
        font-weight: 600;
        text-decoration: none;
        transition:
          background-color var(--mb-transition),
          border-color var(--mb-transition);
      }

      a:hover {
        background-color: var(--mb-color-surface);
        background-image: linear-gradient(var(--mb-color-hover), var(--mb-color-hover));
        border-color: var(--mb-color-border-hover);
      }

      a:active {
        background-color: var(--mb-color-bg);
        background-image: none;
        border-color: var(--mb-color-border-hover);
      }

      span.disabled {
        opacity: 0.55;
        cursor: not-allowed;
        border-color: var(--mb-color-border);
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
    `}};rt([n({attribute:"prev-url"})],N.prototype,"prevUrl");rt([n({attribute:"next-url"})],N.prototype,"nextUrl");rt([n({type:Boolean,attribute:"prev-disabled"})],N.prototype,"prevDisabled");rt([n({type:Boolean,attribute:"next-disabled"})],N.prototype,"nextDisabled");rt([n()],N.prototype,"status");rt([n({attribute:"prev-label"})],N.prototype,"prevLabel");rt([n({attribute:"next-label"})],N.prototype,"nextLabel");rt([n()],N.prototype,"label");b("mb-pagination",N);var Ss=Object.defineProperty,Ct=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Ss(t,e,s),s},it=class extends p{constructor(){super(...arguments),this.open=!1,this.variant="info",this.autoDismiss=4e3,this.message="",this.dismissLabel="Dismiss",this.#t=0,this.#e=t=>{let e=t.detail;e&&(e.variant&&(this.variant=e.variant),e.message!=null&&(this.message=e.message),e.autoDismiss!=null&&(this.autoDismiss=e.autoDismiss),this.show())}}static{this.styles=[u,m`
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
        border: 1px solid var(--mb-color-info);
        border-inline-start-width: 4px;
        background: var(--mb-color-info-soft);
        box-shadow: var(--mb-shadow);
        color: var(--mb-color-fg);
      }

      :host([variant='success']) .toast {
        border-color: var(--mb-color-success);
        background: var(--mb-color-success-soft);
      }

      :host([variant='danger']) .toast {
        border-color: var(--mb-color-danger);
        background: var(--mb-color-danger-soft);
      }

      :host([variant='info']) .toast {
        border-color: var(--mb-color-info);
        background: var(--mb-color-info-soft);
      }

      .message {
        flex: 1;
        font-size: var(--mb-font-size-sm);
        font-weight: 600;
      }

      button {
        appearance: none;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        flex: none;
        min-inline-size: 1.5rem;
        min-block-size: 1.5rem;
        margin-block-start: -0.15rem;
        border: 0;
        border-radius: var(--mb-radius-sm);
        background: transparent;
        color: inherit;
        cursor: pointer;
        font: inherit;
        font-weight: 700;
        line-height: 1;
        padding: 0;
        transition: background-color var(--mb-transition);
      }

      button:hover {
        background: var(--mb-color-hover);
      }

      button:active {
        background: var(--mb-color-border);
      }
    `]}#t;#e;connectedCallback(){super.connectedCallback(),document.addEventListener("mb-toast",this.#e),this.open&&this.#r()}disconnectedCallback(){super.disconnectedCallback(),document.removeEventListener("mb-toast",this.#e),this.#s()}updated(t){(t.has("open")||t.has("autoDismiss"))&&(this.open?this.#r():this.#s())}show(t,e){t!=null&&(this.message=t),e&&(this.variant=e),this.open=!0,this.isConnected&&this.#r()}hide(){this.open=!1}#r(){this.#s(),this.autoDismiss>0&&(this.#t=window.setTimeout(()=>this.hide(),this.autoDismiss))}#s(){this.#t&&(window.clearTimeout(this.#t),this.#t=0)}#i(){this.hide(),this.dispatchEvent(new CustomEvent("mb-close",{bubbles:!0,composed:!0}))}render(){let t=this.variant==="danger"?"alert":"status";return l`
      <div
        part="toast"
        class="toast"
        role=${t}
        aria-live=${this.variant==="danger"?"assertive":"polite"}
        ?hidden=${!this.open}
      >
        <div part="message" class="message">${this.message}<slot></slot></div>
        <button
          type="button"
          part="close"
          aria-label=${this.dismissLabel}
          @click=${this.#i}
        >
          ×
        </button>
      </div>
    `}};Ct([n({type:Boolean,reflect:!0})],it.prototype,"open");Ct([n({reflect:!0})],it.prototype,"variant");Ct([n({type:Number,attribute:"auto-dismiss"})],it.prototype,"autoDismiss");Ct([n()],it.prototype,"message");Ct([n({attribute:"dismiss-label"})],it.prototype,"dismissLabel");b("mb-toast",it);var Es=Object.defineProperty,zt=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Es(t,e,s),s},ot=class extends p{constructor(){super(...arguments),this.value="",this.label="",this.disabled=!1,this.checked=!1,this.name=""}static{this.styles=[u,m`
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
        flex: none;
        margin-block-start: 0;
        accent-color: var(--mb-color-accent);
        inline-size: 1.5rem;
        block-size: 1.5rem;
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
    `}};zt([n()],ot.prototype,"value");zt([n()],ot.prototype,"label");zt([n({type:Boolean,reflect:!0})],ot.prototype,"disabled");zt([n({type:Boolean,reflect:!0})],ot.prototype,"checked");zt([n()],ot.prototype,"name");b("mb-radio",ot);var Ms=Object.defineProperty,H=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Ms(t,e,s),s};function Ls(r){return U(r,t=>!!t&&typeof t=="object"&&typeof t.value=="string"&&typeof t.label=="string",t=>({value:t.value,label:t.label,disabled:!!t.disabled}))}var q=class extends p{constructor(){super(...arguments),this.label="",this.error="",this.value="",this.name="",this.disabled=!1,this.required=!1,this.invalid=!1,this.missingMessage="Please select an option.",this.invalidMessage="Please select a valid option.",this.options=[],this.#t=new P(this),this.#e=new WeakMap,this.#l=t=>{let e=t.detail?.value;e!=null&&(this.#t.markTouched(),this.value=e,this.#o(),this.#a(),this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value},bubbles:!0,composed:!0})))},this.#n=t=>{if(!["ArrowDown","ArrowUp","ArrowRight","ArrowLeft"].includes(t.key))return;let e=this.#i().filter(a=>!a.disabled);if(!e.length)return;t.preventDefault();let i=e.findIndex(a=>a.value===this.value),s=t.key==="ArrowDown"||t.key==="ArrowRight"?1:-1,o=e[(i+s+e.length)%e.length];this.#t.markTouched(),this.value=o.value,this.#o(),this.#a(),o.focus(),this.dispatchEvent(new CustomEvent("mb-change",{detail:{value:this.value},bubbles:!0,composed:!0}))}}static{this.formAssociated=!0}static{this.styles=[u,m`
      :host {
        display: block;
        inline-size: 100%;
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
        gap: var(--mb-space-3);
      }

      .error {
        margin: var(--mb-space-2) 0 0;
        color: var(--mb-color-danger);
        font-size: var(--mb-font-size-sm);
      }
    `]}#t;#e;get#r(){return this.#t.isDisabled}checkValidity(){return this.#t.checkValidity()}reportValidity(){return this.#t.reportValidity()}connectedCallback(){super.connectedCallback(),this.#t.captureDefault(this.value),this.addEventListener("mb-radio-select",this.#l),this.addEventListener("keydown",this.#n)}disconnectedCallback(){super.disconnectedCallback(),this.removeEventListener("mb-radio-select",this.#l),this.removeEventListener("keydown",this.#n)}firstUpdated(){this.#o(),this.#a()}updated(t){(t.has("value")||t.has("name")||t.has("disabled")||t.has("options"))&&this.#o(),(t.has("value")||t.has("required")||t.has("error")||t.has("name")||t.has("disabled")||t.has("missingMessage")||t.has("invalidMessage"))&&this.#a()}formDisabledCallback(t){this.#t.formDisabledCallback(t),this.#o()}formResetCallback(){this.#t.resetInteraction(),this.value=this.#t.defaultValue,this.error="",this.invalid=!1,this.#o(),this.#a()}formStateRestoreCallback(t,e){typeof t=="string"&&(this.value=t)}#s(){return this.renderRoot.querySelector("slot")?.assignedElements({flatten:!0}).filter(t=>t.localName==="mb-radio")??[]}#i(){let t=[...this.renderRoot.querySelectorAll(".options > mb-radio")];return[...this.#s(),...t]}#o(){let t=this.#i();for(let e of t)e.name=this.name||"mb-radio-group",e.checked=e.value===this.value;for(let e of this.#s())this.#e.has(e)||this.#e.set(e,e.disabled),e.disabled=this.#r||!!this.#e.get(e)}#a(){let t=!this.value||this.#i().some(s=>s.value===this.value&&(!s.disabled||this.#r));z(this.#t.internals,this.name&&t?this.value:null);let e=this.required&&!this.value,i=!!this.value&&!t;this.invalid=this.#t.applyConstraintValidity(this.error,e,void 0,this.missingMessage,i?{flags:{badInput:!0},message:this.invalidMessage}:null)}#l;#n;#d(){this.#o()}render(){return l`
      <fieldset part="fieldset" ?disabled=${this.#r}>
        ${this.label?l`<legend part="legend">${this.label}</legend>`:c}
        <div class="options" part="options" role="radiogroup" aria-invalid=${this.invalid?"true":"false"}>
          <slot @slotchange=${this.#d}></slot>
          ${this.options.map(t=>l`
              <mb-radio
                .value=${t.value}
                .label=${t.label}
                ?disabled=${!!t.disabled||this.#r}
                ?checked=${t.value===this.value}
                .name=${this.name||"mb-radio-group"}
              ></mb-radio>
            `)}
        </div>
        ${this.error?l`<p class="error" role="alert">${this.error}</p>`:c}
      </fieldset>
    `}};H([n()],q.prototype,"label");H([n()],q.prototype,"error");H([n()],q.prototype,"value");H([n({reflect:!0})],q.prototype,"name");H([n({type:Boolean,reflect:!0})],q.prototype,"disabled");H([n({type:Boolean,reflect:!0})],q.prototype,"required");H([n({type:Boolean,reflect:!0})],q.prototype,"invalid");H([n({attribute:"missing-message"})],q.prototype,"missingMessage");H([n({attribute:"invalid-message"})],q.prototype,"invalidMessage");H([n({attribute:"options",converter:I(Ls)})],q.prototype,"options");b("mb-radio-group",q);var ce=class extends tt{connectedCallback(){this.hasAttribute("size")||(this.size="md"),super.connectedCallback()}};b("mb-tag",ce);var Ds=Object.defineProperty,Ve=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Ds(t,e,s),s};function Ps(r){return U(r,t=>!!t&&typeof t=="object"&&typeof t.label=="string",t=>({label:t.label,href:t.href,current:!!t.current}))}var St=class extends p{constructor(){super(...arguments),this.label="Breadcrumb",this.items=[]}static{this.styles=[u,m`
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
        text-decoration: underline;
        text-decoration-thickness: 1px;
        text-underline-offset: 0.18em;
        overflow-wrap: anywhere;
      }

      a:hover {
        color: var(--mb-color-accent-hover);
      }

      [aria-current='page'] {
        color: var(--mb-color-muted);
        font-weight: 600;
      }
    `]}render(){return l`
      <nav part="nav" aria-label=${this.label}>
        <ol part="list">
          ${this.items.length?Y(this.items,t=>`${t.href??""}:${t.label}`,t=>l`
                  <li part="item">
                    ${t.current||!t.href?l`<span aria-current=${t.current?"page":c}
                          >${t.label}</span
                        >`:l`<a href=${t.href}>${t.label}</a>`}
                  </li>
                `):l`<slot></slot>`}
        </ol>
      </nav>
    `}};Ve([n()],St.prototype,"label");Ve([n({attribute:"items",converter:I(Ps)})],St.prototype,"items");b("mb-breadcrumbs",St);var Os=Object.defineProperty,Ne=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Os(t,e,s),s},Et=class extends p{constructor(){super(...arguments),this.label="Primary",this.open=!1}static{this.styles=[u,m`
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
        display: inline-flex;
        align-items: center;
        min-block-size: 2rem;
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
        background: var(--mb-color-hover);
      }

      ::slotted(a[aria-current='page']),
      ::slotted(a.is-active),
      ::slotted(a[aria-current='page']:hover),
      ::slotted(a.is-active:hover) {
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
    `}};Ne([n()],Et.prototype,"label");Ne([n({type:Boolean,reflect:!0})],Et.prototype,"open");b("mb-nav",Et);var Rs=Object.defineProperty,Jt=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Rs(t,e,s),s},dt=class extends p{constructor(){super(...arguments),this.expanded=!1,this.for="",this.labelOpen="Menu",this.labelClose="Close menu",this.#t=null,this.#e=new MutationObserver(()=>this.#i())}static{this.styles=[u,m`
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
        border: 1px solid var(--mb-color-border-strong);
        border-radius: var(--mb-radius-md);
        background: var(--mb-color-surface);
        color: var(--mb-color-fg);
        cursor: pointer;
        font: inherit;
        font-weight: 700;
        transition:
          background-color var(--mb-transition),
          border-color var(--mb-transition);
      }

      button:hover {
        background-color: var(--mb-color-surface);
        background-image: linear-gradient(var(--mb-color-hover), var(--mb-color-hover));
        border-color: var(--mb-color-border-hover);
      }

      button:active {
        background-color: var(--mb-color-bg);
        background-image: none;
        border-color: var(--mb-color-border-hover);
      }
    `]}#t;#e;firstUpdated(){this.#s()}updated(t){t.has("for")?this.#s():t.has("expanded")&&this.#o()}disconnectedCallback(){this.#e.disconnect(),super.disconnectedCallback()}#r(){if(!this.for)return null;let t=this.getRootNode();return("getElementById"in t?t.getElementById(this.for):null)??this.ownerDocument.getElementById(this.for)}#s(){this.#e.disconnect(),this.#t=this.#r(),this.#t&&(this.#i(),this.#e.observe(this.#t,{attributes:!0,attributeFilter:["open"]}))}#i(){let t=this.#t;if(!t)return;let e="open"in t?!!t.open:t.hasAttribute("open");this.expanded!==e&&(this.expanded=e)}#o(){let t=this.#t;t&&(t.toggleAttribute("open",this.expanded),"open"in t&&(t.open=this.expanded))}#a(){this.expanded=!this.expanded,this.#o(),this.dispatchEvent(new CustomEvent("mb-toggle",{detail:{expanded:this.expanded},bubbles:!0,composed:!0}))}render(){return l`
      <button
        part="button"
        type="button"
        aria-expanded=${this.expanded?"true":"false"}
        aria-controls=${this.for||c}
        aria-label=${this.expanded?this.labelClose:this.labelOpen}
        @click=${this.#a}
      >
        <slot>${this.expanded?"\u2715":"\u2630"}</slot>
      </button>
    `}};Jt([n({type:Boolean,reflect:!0})],dt.prototype,"expanded");Jt([n({attribute:"for"})],dt.prototype,"for");Jt([n({attribute:"label-open"})],dt.prototype,"labelOpen");Jt([n({attribute:"label-close"})],dt.prototype,"labelClose");b("mb-nav-toggle",dt);var qs=Object.defineProperty,ft=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&qs(t,e,s),s},G=class extends p{constructor(){super(...arguments),this.src="",this.alt="",this.name="",this.size="md",this.fallbackLabel="Avatar",this._failed=!1}static{this.styles=[u,m`
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
        box-shadow: inset 0 0 0 1px var(--mb-color-border-strong);
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
      <span part="base" class="avatar" role=${t?c:"img"} aria-label=${t?c:this.alt||this.name||this.fallbackLabel}>
        ${t?l`<img part="image" src=${this.src} alt=${this.alt} @error=${this.#e} />`:l`<span part="initials">${this.#t}</span>`}
      </span>
    `}};ft([n({reflect:!0})],G.prototype,"src");ft([n()],G.prototype,"alt");ft([n()],G.prototype,"name");ft([n({reflect:!0})],G.prototype,"size");ft([n({attribute:"fallback-label"})],G.prototype,"fallbackLabel");ft([B()],G.prototype,"_failed");b("mb-avatar",G);var Ts=Object.defineProperty,je=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Ts(t,e,s),s},Mt=class extends p{constructor(){super(...arguments),this.size="md",this.label="Loading"}static{this.styles=[u,m`
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
    `}};je([n({reflect:!0})],Mt.prototype,"size");je([n()],Mt.prototype,"label");b("mb-spinner",Mt);var he=class extends p{static{this.styles=[u,m`
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
        /* Shared control track so slotted fields / buttons / filters match height. */
        --mb-control-height: 2.5rem;
        --mb-control-height-sm: 2.5rem;
      }

      .end {
        margin-inline-start: auto;
      }

      /*
        Fields default to inline-size: 100% (form stacks). In a toolbar row,
        let them share the track instead of forcing a full-width wrap.
      */
      ::slotted(mb-input),
      ::slotted(mb-select),
      ::slotted(mb-combobox),
      ::slotted(mb-textarea) {
        flex: 1 1 12rem;
        inline-size: auto;
        max-inline-size: 20rem;
      }

      ::slotted(mb-segmented-control) {
        flex: 0 1 auto;
        max-inline-size: 100%;
      }

      /* Stretch interactive chrome across the row on narrow viewports. */
      @media (max-width: 36rem) {
        .toolbar {
          flex-direction: column;
          align-items: stretch;
        }

        .start,
        .end {
          flex-direction: column;
          align-items: stretch;
          inline-size: 100%;
          margin-inline-start: 0;
        }

        ::slotted(*),
        ::slotted(mb-input),
        ::slotted(mb-select),
        ::slotted(mb-combobox),
        ::slotted(mb-textarea),
        ::slotted(mb-button),
        ::slotted(mb-segmented-control) {
          flex: 1 1 auto;
          inline-size: 100%;
          max-inline-size: none;
        }
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
    `}};b("mb-toolbar",he);var Ue="(max-width: 36rem)";function Bs(r){return!!r&&typeof r=="object"&&typeof r.id=="string"&&typeof r.label=="string"}function Vs(r){return{id:r.id,label:r.label,collapsed:!!r.collapsed,meta:typeof r.meta=="string"?r.meta:void 0,count:r.count===!1?!1:void 0}}function Ns(r){return U(r,Bs,Vs)}var Ie=I(Ns);function He(r,t){let e=Number(r),i=Number(t);return r!==""&&t!==""&&!Number.isNaN(e)&&!Number.isNaN(i)?e-i:r.localeCompare(t,void 0,{sensitivity:"base",numeric:!0})}var Fe=m`
  :host {
    display: block;
    inline-size: 100%;
    max-inline-size: 100%;
    --mb-table-template: repeat(var(--mb-table-col-count, 1), minmax(0, 1fr));
  }

  .root {
    display: flex;
    flex-direction: column;
    gap: var(--mb-space-3);
    inline-size: 100%;
    max-inline-size: 100%;
  }

  .caption {
    margin: 0;
    font-family: var(--mb-font-display);
    font-size: var(--mb-font-size-lg);
    font-weight: 600;
    line-height: var(--mb-line-height-tight);
  }

  .frame {
    display: flex;
    flex-direction: column;
    gap: var(--mb-space-3);
    inline-size: 100%;
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
    min-block-size: var(--mb-control-height, 2.5rem);
    border: 1px solid var(--mb-color-border-strong);
    border-radius: var(--mb-radius-md);
    background: var(--mb-color-bg);
    transition:
      background-color var(--mb-transition),
      border-color var(--mb-transition);
    color: var(--mb-color-fg);
    font: inherit;
    font-family: var(--mb-font-display);
    font-weight: 600;
    text-align: start;
    cursor: pointer;
  }

  .section-head:hover {
    background-image: linear-gradient(var(--mb-color-hover), var(--mb-color-hover));
  }

  .section-head:active {
    background-image: linear-gradient(var(--mb-color-border), var(--mb-color-border));
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
    border: 1px solid var(--mb-color-border-strong);
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
`,Ke=m`
  :host {
    display: block;
    inline-size: 100%;
    min-inline-size: 0;
  }

  .wrap {
    display: flex;
    align-items: stretch;
    gap: var(--mb-space-2);
    inline-size: 100%;
    min-inline-size: 0;
  }

  .handle,
  .spacer {
    flex: none;
    inline-size: 1.5rem;
    block-size: 1.5rem;
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

  .handle:hover {
    background: var(--mb-color-hover);
    color: var(--mb-color-fg);
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
    inline-size: 100%;
    min-inline-size: 0;
    flex: 1 1 auto;
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
    font-weight: 600;
    color: var(--mb-color-muted);
    background: transparent;
    padding-block: var(--mb-space-2);
  }

  :host([data-mode='cards']) .wrap {
    padding-block: var(--mb-space-4);
    padding-inline: var(--mb-space-4);
    background: var(--mb-color-surface);
    border: 1px solid var(--mb-color-border-strong);
    border-radius: var(--mb-radius-lg);
    box-shadow: var(--mb-shadow-sm);
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
`,Je=m`
  :host {
    display: block;
    inline-size: 100%;
    min-inline-size: 0;
  }

  .cell {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: var(--mb-space-1);
    inline-size: 100%;
    min-inline-size: 0;
  }

  .label {
    display: none;
    font-size: var(--mb-font-size-sm);
    font-weight: 600;
    color: var(--mb-color-muted);
  }

  .value {
    inline-size: 100%;
    min-inline-size: 0;
    max-inline-size: 100%;
  }

  .sort {
    display: inline-flex;
    align-items: center;
    gap: var(--mb-space-1);
    max-inline-size: 100%;
    min-block-size: 1.5rem;
    margin: 0;
    padding-block: 0.125rem;
    padding-inline: 0.125rem;
    border: none;
    border-radius: var(--mb-radius-sm);
    background: transparent;
    color: inherit;
    font: inherit;
    font-weight: inherit;
    text-align: inherit;
    cursor: pointer;
  }

  .sort:hover {
    color: var(--mb-color-fg);
    background: var(--mb-color-hover);
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

  :host([align='center']) .value {
    display: flex;
    justify-content: center;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--mb-space-2);
  }

  :host([align='end']) .cell {
    align-items: stretch;
    text-align: end;
  }

  :host([align='end']) .value {
    display: flex;
    justify-content: flex-end;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--mb-space-2);
  }

  :host([data-mode='cards']) .label:not([hidden]) {
    display: block;
  }

  :host([data-mode='cards'][primary]) .value {
    font-family: var(--mb-font-display);
    font-weight: 600;
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

  /* Actions column: pin controls to the inline-end of the cell track. */
  :host([actions]) .cell {
    align-items: stretch;
    text-align: end;
  }

  :host([actions]) .value,
  :host([actions][data-mode='cards']) .value,
  :host([actions][data-mode='table']) .value {
    display: flex;
    justify-content: flex-end;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--mb-space-2);
    inline-size: 100%;
  }
`;var js=Object.defineProperty,T=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&js(t,e,s),s},M=class extends p{constructor(){super(...arguments),this.label="",this.columns="",this.density="default",this.layout="auto",this.sections=[],this.sortKey="",this.sortDirection="asc",this.reorderable=!1,this.reorderLabel="Drag to reorder",this.sortLabel="Sort by {name}",this.hideCount=!1,this.stickyHeader=!1,this._sectionCounts={},this.#e=()=>this.#m(),this.#r=!1,this.#s=!1,this.#i=!1,this.#o=null,this.#a=null,this.#l="before",this.#n=null,this.#_=t=>{if(!this.#o)return;let e=document.elementsFromPoint(t.clientX,t.clientY),i=e.find(o=>o instanceof HTMLElement&&o.localName==="mb-table-row"&&this.contains(o)&&o!==this.#o&&o.slot!=="head"&&!o.hasAttribute("head")),s=e.find(o=>o instanceof HTMLElement&&this.renderRoot.contains(o)&&o.hasAttribute("data-section"));if(this.#A(),i){let o=i.getBoundingClientRect(),a=t.clientY<o.top+o.height/2?"before":"after";this.#a=i,this.#l=a,this.#n=i.section.trim()||null,i.setAttribute("data-drop",a);return}if(s){let o=s.getAttribute("data-section");o&&(this.#n=o,s.toggleAttribute("data-drop-section",!0))}},this.#v=()=>{let t=this.#o,e=this.#a,i=this.#l,s=this.#n;if(window.removeEventListener("pointermove",this.#_),window.removeEventListener("pointerup",this.#v),window.removeEventListener("pointercancel",this.#v),t?.removeAttribute("data-dragging"),this.#A(),this.#o=null,!!t){if(e){this.moveRow(t,{before:i==="before"?e:void 0,after:i==="after"?e:void 0,section:e.section.trim()||void 0});return}s!=null&&this.moveRow(t,{section:s})}},this.#g=()=>{this.#r||this.#s||queueMicrotask(()=>{this.#r||this.#s||(this.#m(),this.#b())})},this.#z=t=>{let e=t.composedPath(),i=e.find(s=>s instanceof HTMLElement&&s.localName==="mb-table-cell");!i?.sortKey.trim()||t.defaultPrevented||e.some(s=>s instanceof Element&&s.matches("mb-input, mb-select, mb-textarea, mb-button, a, input, select, textarea"))||this.#y(i.sortKey.trim())}}static{this.styles=[u,Fe]}#t;#e;#r;#s;#i;#o;#a;#l;#n;connectedCallback(){super.connectedCallback(),this.setAttribute("role","table"),this.#t=window.matchMedia(Ue),this.#t.addEventListener("change",this.#e),this.#c(),queueMicrotask(()=>this.#m())}disconnectedCallback(){this.#t?.removeEventListener("change",this.#e),super.disconnectedCallback()}updated(t){this.label?this.setAttribute("aria-label",this.label):this.removeAttribute("aria-label");let e=t.has("columns")||t.has("layout")||t.has("density")||t.has("sections")||t.has("reorderable")||t.has("reorderLabel")||t.has("sortLabel"),i=t.has("sortKey")||t.has("sortDirection")||t.has("sections");(e||i)&&queueMicrotask(()=>{e&&(this.#c(),this.#m()),i&&(this.#f(),this.#b())})}get#d(){return this.layout==="table"?"table":this.layout==="cards"||this.#t?.matches?"cards":"table"}get#p(){return this.sections.length>0}#c(){let t=this.columns.trim();if(!t){this.style.removeProperty("--mb-table-template"),this.style.removeProperty("--mb-table-col-count");return}if(/^\d+$/.test(t)){this.style.setProperty("--mb-table-col-count",t),this.style.setProperty("--mb-table-template",`repeat(${t}, minmax(0, 1fr))`);return}this.style.setProperty("--mb-table-template",t)}refreshRows(){this.#i||(this.#m(),this.#b())}#h(){return[...this.querySelectorAll("mb-table-row")].filter(t=>t.slot!=="head"&&!t.hasAttribute("head"))}#$(){return this.querySelector('mb-table-row[slot="head"]')??this.querySelector("mb-table-row[head]")}#S(){let t=new Set(this.sections.map(s=>s.id)),e={};for(let s of this.sections)e[s.id]=0;for(let s of this.#h()){let o=s.section.trim();if(o&&t.has(o)){let a=`section-${o}`;s.slot!==a&&(s.slot=a),e[o]=(e[o]??0)+1}else s.slot.startsWith("section-")&&(s.slot="")}let i=this._sectionCounts;Object.keys(e).length===Object.keys(i).length&&Object.keys(e).every(s=>i[s]===e[s])||(this._sectionCounts=e)}#m(){if(!this.#s){this.#s=!0;try{let t=this.#d;this.setAttribute("data-mode",t),this.#S(),this.querySelectorAll("mb-table-row").forEach(e=>{e.setAttribute("data-mode",t),e.toggleAttribute("data-compact",this.density==="compact");let i=e.slot==="head"||e.hasAttribute("head");e.toggleAttribute("data-reorderable",this.reorderable&&!i),e.toggleAttribute("data-reorder-spacer",this.reorderable&&i),this.reorderable&&!i?e.setAttribute("data-reorder-label",this.reorderLabel):e.removeAttribute("data-reorder-label"),e.requestUpdate()}),this.querySelectorAll("mb-table-cell").forEach(e=>{e.setAttribute("data-mode",t),e.toggleAttribute("data-compact",this.density==="compact"),(e.sortKey.trim()||e.sortable)&&e.setAttribute("data-sort-label",this.sortLabel),e.requestUpdate()}),this.#x(),this.#f(),this.#u()}finally{this.#s=!1}}}#u(){if(!this.#p)return;let t=this.renderRoot.querySelector("slot.ungrouped-slot"),e=this.renderRoot.querySelector(".ungrouped");if(!t||!e)return;let i=t.assignedElements({flatten:!0}).some(s=>s.localName==="mb-table-row");e.toggleAttribute("data-has-content",i)}#x(){let t=this.#$();if(!t)return;let e=[...t.querySelectorAll("mb-table-cell")],i=e.map(s=>s.hideLabel||s.actions?"":(s.textContent??"").replace(/\s+/g," ").trim());if(e.length){this.columns.trim()||(this.style.setProperty("--mb-table-col-count",String(e.length)),this.style.setProperty("--mb-table-template",`repeat(${e.length}, minmax(0, 1fr))`));for(let s of this.#h())[...s.querySelectorAll(":scope > mb-table-cell")].forEach((o,a)=>{if(o.hideLabel||o.actions){o.dataset.labelLocked="true",o.label&&(o.label="");return}if(o.dataset.labelLocked==="true")return;if(o.hasAttribute("label")){o.dataset.labelLocked="true";return}let d=i[a];d&&o.label!==d&&(o.label=d)})}}#f(){let t=this.#$();if(t)for(let e of t.querySelectorAll("mb-table-cell")){let i=e.sortKey.trim(),s=!!i&&i===this.sortKey,o=s?this.sortDirection:null;i&&!e.sortable&&(e.sortable=!0),e.sortActive!==s&&(e.sortActive=s),e.sortDirection!==o&&(e.sortDirection=o)}}#k(t){let e=this.#$();return e?[...e.querySelectorAll("mb-table-cell")].findIndex(i=>i.sortKey.trim()===t):-1}#w(t,e){if(t.sortValue.trim()&&(!e||this.#k(e)<0))return t.sortValue.trim();let i=this.#k(e);if(i<0)return t.sortValue.trim();let s=t.querySelectorAll(":scope > mb-table-cell")[i];if(!s)return"";if(s.sortValue.trim())return s.sortValue.trim();let o=s.querySelector("mb-input, mb-select, mb-textarea, input, select, textarea");return o&&typeof o.value=="string"&&o.value!==""?o.value:(s.textContent??"").replace(/\s+/g," ").trim()}#b(){if(!(this.#r||!this.sortKey.trim())){this.#r=!0;try{let t=this.sortKey.trim(),e=this.sortDirection==="desc"?-1:1,i=this.#p?this.sections.map(s=>this.#h().filter(o=>o.section.trim()===s.id)):[this.#h()];for(let s of i){let o=[...s].sort((a,d)=>e*He(this.#w(a,t),this.#w(d,t)));if(!(s.length===o.length&&s.every((a,d)=>a===o[d])))for(let a of o)this.appendChild(a)}}finally{this.#r=!1}}}#y(t){this.sortKey===t?this.sortDirection=this.sortDirection==="asc"?"desc":"asc":(this.sortKey=t,this.sortDirection="asc"),this.#f(),this.#b(),this.dispatchEvent(new CustomEvent("mb-sort",{detail:{key:this.sortKey,direction:this.sortDirection},bubbles:!0,composed:!0}))}beginReorder(t,e){!this.reorderable||t.head||t.slot==="head"||this.#o||e.button!==void 0&&e.button!==0||(e.preventDefault(),e.stopPropagation(),this.#o=t,t.toggleAttribute("data-dragging",!0),window.addEventListener("pointermove",this.#_),window.addEventListener("pointerup",this.#v),window.addEventListener("pointercancel",this.#v))}#A(){this.querySelectorAll("mb-table-row[data-drop]").forEach(t=>{t.removeAttribute("data-drop")}),this.renderRoot.querySelectorAll("[data-drop-section]").forEach(t=>{t.removeAttribute("data-drop-section")}),this.#a=null,this.#n=null}#_;#v;moveRow(t,e={}){if(t.head||t.slot==="head"||!this.contains(t)||e.before&&!this.contains(e.before)||e.after&&!this.contains(e.after))return;let i=t.section.trim(),s=e.section??i,o=!1;this.#i=!0,this.sortKey&&(this.sortKey="",this.#f());try{if(e.section!=null&&e.section!==i&&(t.section=e.section,s=e.section,o=!0),e.before&&e.before!==t)e.before.previousElementSibling!==t&&(e.before.before(t),o=!0),s=e.before.section.trim()||s,t.section.trim()!==s&&(t.section=s,o=!0);else if(e.after&&e.after!==t)e.after.nextElementSibling!==t&&(e.after.after(t),o=!0),s=e.after.section.trim()||s,t.section.trim()!==s&&(t.section=s,o=!0);else if(e.section!=null){let d=this.#h().filter(v=>v!==t&&v.section.trim()===e.section),h=d[d.length-1];h?(h.after(t),o=!0):(this.appendChild(t),o=!0)}}finally{this.#i=!1}if(!o)return;this.#m();let a=this.#h().map(d=>({id:d.id||d.getAttribute("data-id")||"",section:d.section.trim()}));this.dispatchEvent(new CustomEvent("mb-reorder",{detail:{rowId:t.id||t.getAttribute("data-id")||"",fromSection:i,toSection:s,beforeId:e.before?e.before.id||e.before.getAttribute("data-id")||"":null,afterId:e.after?e.after.id||e.after.getAttribute("data-id")||"":null,order:a},bubbles:!0,composed:!0}))}moveRowByKeyboard(t,e){if(!this.reorderable||!this.contains(t))return;let i=this.#h(),s=i.indexOf(t),o=i[s+e];s<0||!o||(e<0?this.moveRow(t,{before:o,section:o.section||void 0}):this.moveRow(t,{after:o,section:o.section||void 0}))}#E(t){let e=this.sections.findIndex(o=>o.id===t);if(e<0)return;let i=this.sections.map((o,a)=>a===e?{...o,collapsed:!o.collapsed}:o);this.sections=i;let s=!!i[e]?.collapsed;this.dispatchEvent(new CustomEvent("mb-section-toggle",{detail:{id:t,collapsed:s},bubbles:!0,composed:!0}))}#g;#C(t){let e=t.target.assignedNodes({flatten:!0}).length>0;this.renderRoot.querySelector(".empty")?.toggleAttribute("data-has-content",e)}#M(t){return!(this.hideCount||t.count===!1)}#z;render(){return l`
      <div part="root" class="root">
        ${this.label?l`<div part="caption" class="caption">${this.label}</div>`:c}
        <div part="frame" class="frame">
          <div part="head" class="head" @click=${this.#z}>
            <slot name="head" @slotchange=${this.#g}></slot>
          </div>
          <div part="body" class="body">
            ${this.#p?l`
                  ${Y(this.sections,t=>t.id,t=>l`
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
                          @click=${()=>this.#E(t.id)}
                        >
                          <span class="section-label">${t.label}</span>
                          <span class="section-meta">
                            ${t.meta?l`<span part="section-custom-meta">${t.meta}</span>`:c}
                            <slot name=${`section-meta-${t.id}`}></slot>
                            ${this.#M(t)?l`<span part="section-count"
                                  >${this._sectionCounts[t.id]??0}</span
                                >`:c}
                            <span class="section-chevron" aria-hidden="true">▾</span>
                          </span>
                        </button>
                        <div part="section-rows" class="section-rows">
                          <slot
                            name=${`section-${t.id}`}
                            @slotchange=${this.#g}
                          ></slot>
                        </div>
                      </section>
                    `)}
                  <div part="ungrouped" class="ungrouped section">
                    <slot
                      class="ungrouped-slot"
                      @slotchange=${this.#g}
                    ></slot>
                  </div>
                `:l`<slot @slotchange=${this.#g}></slot>`}
          </div>
        </div>
        <div part="empty" class="empty">
          <slot name="empty" @slotchange=${this.#C}></slot>
        </div>
      </div>
    `}};T([n()],M.prototype,"label");T([n()],M.prototype,"columns");T([n({reflect:!0})],M.prototype,"density");T([n({reflect:!0})],M.prototype,"layout");T([n({attribute:"sections",converter:Ie})],M.prototype,"sections");T([n({attribute:"sort-key",reflect:!0})],M.prototype,"sortKey");T([n({attribute:"sort-direction",reflect:!0})],M.prototype,"sortDirection");T([n({type:Boolean,reflect:!0})],M.prototype,"reorderable");T([n({attribute:"reorder-label"})],M.prototype,"reorderLabel");T([n({attribute:"sort-label"})],M.prototype,"sortLabel");T([n({type:Boolean,reflect:!0,attribute:"hide-count"})],M.prototype,"hideCount");T([n({type:Boolean,reflect:!0,attribute:"sticky-header"})],M.prototype,"stickyHeader");T([B()],M.prototype,"_sectionCounts");b("mb-table",M);var Us=Object.defineProperty,de=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Us(t,e,s),s},pt=class extends p{constructor(){super(...arguments),this.head=!1,this.section="",this.sortValue="",this.#t=t=>{this.closest("mb-table")?.beginReorder(this,t)},this.#e=t=>{t.key!=="ArrowUp"&&t.key!=="ArrowDown"||(t.preventDefault(),this.closest("mb-table")?.moveRowByKeyboard(this,t.key==="ArrowUp"?-1:1))}}static{this.styles=[u,Ke]}connectedCallback(){super.connectedCallback(),this.setAttribute("role","row"),this.head&&this.slot!=="head"&&(this.slot="head")}attributeChangedCallback(t,e,i){super.attributeChangedCallback(t,e,i),t==="data-reorder-label"&&e!==i&&this.requestUpdate()}updated(t){if(t.has("head")&&this.head&&(this.slot="head"),t.has("section")){let e=t.get("section");if(e!==void 0||this.section){let i=this.closest("mb-table");i&&e!==void 0&&i.refreshRows()}}(this.slot==="head"||this.head)&&this.getAttribute("data-mode")==="cards"?this.setAttribute("aria-hidden","true"):this.removeAttribute("aria-hidden")}#t;#e;#r(){return this.getAttribute("data-reorder-label")?.trim()||this.closest("mb-table")?.reorderLabel?.trim()||"Drag to reorder"}render(){let t=this.hasAttribute("data-reorderable"),e=this.hasAttribute("data-reorder-spacer");return l`
      <div part="wrap" class="wrap">
        ${t?l`
              <button
                type="button"
                part="handle"
                class="handle"
                aria-label=${this.#r()}
                aria-keyshortcuts="ArrowUp ArrowDown"
                @pointerdown=${this.#t}
                @keydown=${this.#e}
              >
                ⠿
              </button>
            `:e?l`<span class="spacer" aria-hidden="true"></span>`:c}
        <div part="row" class="row">
          <slot></slot>
        </div>
      </div>
    `}};de([n({type:Boolean,reflect:!0})],pt.prototype,"head");de([n({reflect:!0})],pt.prototype,"section");de([n({attribute:"sort-value"})],pt.prototype,"sortValue");b("mb-table-row",pt);var Is=Object.defineProperty,F=(r,t,e,i)=>{for(var s=void 0,o=r.length-1,a;o>=0;o--)(a=r[o])&&(s=a(t,e,s)||s);return s&&Is(t,e,s),s},O=class extends p{constructor(){super(...arguments),this.label="",this.align="start",this.primary=!1,this.hideLabel=!1,this.actions=!1,this.sortKey="",this.sortable=!1,this.sortValue="",this.sortActive=!1,this.sortDirection=null}static{this.styles=[u,Je]}connectedCallback(){super.connectedCallback(),this.#e()}attributeChangedCallback(t,e,i){super.attributeChangedCallback(t,e,i),t==="data-sort-label"&&e!==i&&this.requestUpdate()}updated(t){if(this.#e(),t.has("sortKey")&&this.sortKey.trim()&&(this.sortable=!0),(t.has("hideLabel")||t.has("actions"))&&(this.hideLabel||this.actions)&&(this.dataset.labelLocked="true",this.label&&(this.label="")),this.#t()&&this.sortKey.trim()){let e=this.sortActive&&this.sortDirection?this.sortDirection:"none";this.setAttribute("aria-sort",e)}else this.removeAttribute("aria-sort")}#t(){let t=this.parentElement;return t?.slot==="head"||t?.hasAttribute("head")===!0}#e(){this.setAttribute("role",this.#t()?"columnheader":"cell")}#r(){return!this.sortActive||!this.sortDirection?"\u2195":this.sortDirection==="asc"?"\u2191":"\u2193"}#s(){let t=this.sortKey.trim()||this.label.trim()||(this.textContent??"").replace(/\s+/g," ").trim()||"column",e=this.getAttribute("data-sort-label")?.trim()||this.closest("mb-table")?.sortLabel?.trim()||"Sort by {name}";return e.includes("{name}")?e.replace(/\{name\}/g,t):`${e} ${t}`.trim()}render(){let t=!!this.label&&this.getAttribute("data-mode")==="cards"&&!this.#t()&&!this.hideLabel&&!this.actions,e=this.#t()&&(this.sortable||!!this.sortKey.trim());return l`
      <div part="cell" class="cell">
        <span part="label" class="label" ?hidden=${!t}>${this.label}</span>
        <div part="value" class="value">
          ${e?l`
                <button
                  type="button"
                  part="sort"
                  class="sort"
                  aria-label=${this.#s()}
                >
                  <slot></slot>
                  <span class="sort-indicator" aria-hidden="true">${this.#r()}</span>
                </button>
              `:l`<slot></slot>`}
        </div>
      </div>
    `}};F([n()],O.prototype,"label");F([n({reflect:!0})],O.prototype,"align");F([n({type:Boolean,reflect:!0})],O.prototype,"primary");F([n({type:Boolean,reflect:!0,attribute:"hide-label"})],O.prototype,"hideLabel");F([n({type:Boolean,reflect:!0})],O.prototype,"actions");F([n({attribute:"sort-key",reflect:!0})],O.prototype,"sortKey");F([n({type:Boolean,reflect:!0})],O.prototype,"sortable");F([n({attribute:"sort-value"})],O.prototype,"sortValue");F([n({type:Boolean,reflect:!0,attribute:"sort-active"})],O.prototype,"sortActive");F([n({attribute:!1})],O.prototype,"sortDirection");b("mb-table-cell",O);
