// GOcms 主题切换
function setTheme(name){
  document.getElementById('theme-css').href='/static/css/theme-'+name+'.css';
  localStorage.setItem('gocms-theme',name);
}
(function(){
  var t=localStorage.getItem('gocms-theme')||'red';
  var link=document.createElement('link');
  link.id='theme-css';
  link.rel='stylesheet';
  link.href='/static/css/theme-'+t+'.css';
  document.head.appendChild(link);
})();
