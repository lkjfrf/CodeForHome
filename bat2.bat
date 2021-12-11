for %%f in (Videos\*) do (mp4box -raw 2 %%f -out audio.aac
if exist audio.aac (
del audio.aac
MP4Box -dash 2000 -profile dashavc264:live -bs-switching multi -url-template Videos/%%~nf.mp4#trackID=1:id=vid0:role=vid0 Videos/%%~nf.mp4#trackID=2:id=aud0:role=aud0 -out %%~nf/%%~nf.mpd
) else (
MP4Box -dash 2000 -profile dashavc264:live -out %%~nf/%%~nf.mpd Videos/%%~nf.mp4
)
)