import { useState, useRef } from "react";
import MatchResult, { type Match } from "./MatchResult";

const AudioRecorder = () => {
  const [isRecording, setIsRecording] = useState(false);
  const [isAnalyzing, setIsAnalyzing] = useState(false);
  const [matchResult, setMatchResult] = useState<Match[]>([]);
  const [showCatalog, setShowCatalog] = useState(false);
  const [songsList, setSongsList] = useState<Match[]>([]);
  const [isLoadingSongs, setIsLoadingSongs] = useState(false);
  const [audioSource, setAudioSource] = useState<"mic" | "system">("system");

  const fetchSongs = async () => {
    setIsLoadingSongs(true);
    setShowCatalog(true);
    try {
      const res = await fetch("/api/v1/songs");
      if (res.ok) {
        const data = await res.json();
        setSongsList(data.songs || []);
      }
    } catch (e) {
      console.error("Failed to load catalog", e);
    } finally {
      setIsLoadingSongs(false);
    }
  };

  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const chunksRef = useRef<Blob[]>([]);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const retryCountRef = useRef(0);

  const startRecording = async () => {
    try {
      let stream: MediaStream;

      if (audioSource === "system") {
        if (!navigator.mediaDevices.getDisplayMedia) {
          alert("توهان جو برائوزر سسٽم آڊيو سپورٽ نٿو ڪري. مهرباني ڪري مائيڪرو فون استعمال ڪريو.");
          return;
        }

        const displayStream = await navigator.mediaDevices.getDisplayMedia({
          video: true,
          audio: {
            echoCancellation: false,
            noiseSuppression: false,
            autoGainControl: false,
          },
        });

        const audioTracks = displayStream.getAudioTracks();
        if (audioTracks.length === 0) {
          displayStream.getTracks().forEach((t) => t.stop());
          alert("مهرباني ڪري 'Also share system audio' يا 'Tab Audio' چيڪ ڪريو ته جيئن ڪمپيوٽر جو آواز حاصل ٿئي.");
          setIsRecording(false);
          setIsAnalyzing(false);
          return;
        }

        // Stop video track immediately since we only need the audio
        displayStream.getVideoTracks().forEach((track) => track.stop());
        stream = new MediaStream(audioTracks);
      } else {
        stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      }

      mediaRecorderRef.current = new MediaRecorder(stream);

      mediaRecorderRef.current.ondataavailable = (e) => {
        if (e.data.size > 0) {
          chunksRef.current.push(e.data);
        }
      };

      mediaRecorderRef.current.onstop = () => {
        const blob = new Blob(chunksRef.current, { type: "audio/webm" });
        const reader = new FileReader();
        reader.onloadend = async () => {
          const base64 = reader.result?.toString().split(",")[1];
          if (base64) {
            await sendToBackend(base64);
          } else {
            handleRetryOrFailure();
          }
        };
        reader.readAsDataURL(blob);
        chunksRef.current = [];
      };

      mediaRecorderRef.current.start();
      setIsRecording(true);
      setIsAnalyzing(false);

      timeoutRef.current = setTimeout(() => {
        stopRecording();
      }, 3500);
    } catch (error) {
      console.error("Error accessing audio source:", error);
      setIsRecording(false);
      setIsAnalyzing(false);
    }
  };

  const stopRecording = () => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
    if (
      mediaRecorderRef.current &&
      mediaRecorderRef.current.state === "recording"
    ) {
      mediaRecorderRef.current.stop();
      setIsRecording(false);
      setIsAnalyzing(true);

      mediaRecorderRef.current.stream
        .getTracks()
        .forEach((track) => track.stop());
    }
  };

  const handleRetryOrFailure = () => {
    if (retryCountRef.current < 2) {
      retryCountRef.current += 1;
      startRecording();
    } else {
      retryCountRef.current = 0;
      setIsAnalyzing(false);
      alert("ڪو به گيت نه مليو. مهرباني ڪري ٻيهر ڪوشش ڪريو.");
    }
  };

  const sendToBackend = async (base64Data: string) => {
    try {
      const response = await fetch("/api/v1/songs/match", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ data: base64Data }),
      });

      if (!response.ok) {
        throw new Error("Failed to match audio");
      }

      const result = await response.json();
      const matches = result.records || [];

      if (matches.length > 0) {
        setMatchResult(matches);
        retryCountRef.current = 0;
        setIsAnalyzing(false);
      } else {
        handleRetryOrFailure();
      }
    } catch (error) {
      console.error("Error sending audio to backend:", error);
      handleRetryOrFailure();
    }
  };

  const handleReset = () => {
    setMatchResult([]);
    setIsRecording(false);
    setIsAnalyzing(false);
    retryCountRef.current = 0;
  };

  const isBusy = isRecording || isAnalyzing;

  if (matchResult.length > 0) {
    return <MatchResult match={matchResult[0]} onReset={handleReset} />;
  }

  return (
    <div className="h-screen overflow-hidden bg-[#0474ff] text-white p-4 relative" dir="rtl">
      {/* Top Bar with Catalog Button */}
      <div className="absolute top-4 left-4 right-4 flex justify-between items-center z-20">
        <div className="text-sm font-medium bg-white/20 backdrop-blur-md px-3 py-1.5 rounded-full flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
          ڊيٽابيس: فعال (MySQL)
        </div>
        <button
          onClick={fetchSongs}
          className="bg-white text-[#0474ff] hover:bg-slate-100 font-bold px-4 py-1.5 rounded-full shadow-md text-sm transition-all"
        >
          گيتن جي فهرست (Library)
        </button>
      </div>

      <div className="flex flex-col items-center justify-center h-full gap-6">
        <div className="text-center space-y-2">
          <h3
            className={`text-3xl font-bold tracking-tight transition-all duration-300 ${
              isBusy ? "opacity-0 translate-y-4" : "opacity-100 translate-y-0"
            }`}
          >
            هيلو سنڌ ميوزڪ سڃاڻپ
          </h3>

          {/* Audio Source Switcher */}
          {!isBusy && (
            <div className="flex items-center justify-center gap-1.5 bg-black/25 p-1 rounded-full border border-white/20 mx-auto max-w-xs transition-all">
              <button
                type="button"
                onClick={() => setAudioSource("system")}
                className={`px-3.5 py-1.5 rounded-full text-xs font-bold transition-all flex items-center gap-1.5 ${
                  audioSource === "system"
                    ? "bg-white text-[#0474ff] shadow-md"
                    : "text-white/80 hover:text-white"
                }`}
              >
                <span>💻</span> سسٽم / ڪمپيوٽر آڊيو
              </button>
              <button
                type="button"
                onClick={() => setAudioSource("mic")}
                className={`px-3.5 py-1.5 rounded-full text-xs font-bold transition-all flex items-center gap-1.5 ${
                  audioSource === "mic"
                    ? "bg-white text-[#0474ff] shadow-md"
                    : "text-white/80 hover:text-white"
                }`}
              >
                <span>🎙️</span> مائيڪرو فون
              </button>
            </div>
          )}
        </div>

        <div className="relative flex items-center justify-center">
          {isRecording && (
            <>
              <div className="absolute inset-0 rounded-full bg-white/30 animate-ripple" />
              <div className="absolute inset-0 rounded-full bg-white/30 animate-ripple [animation-delay:0.8s]" />
              <div className="absolute inset-0 rounded-full bg-white/30 animate-ripple [animation-delay:1.6s]" />
            </>
          )}
          {isAnalyzing && (
            <div className="absolute inset-0 rounded-full border-4 border-white/30 border-t-white animate-spin" />
          )}

          <button
            className={`relative z-10 w-32 h-32 rounded-full border-none bg-white flex flex-col items-center justify-center animate-in-out shadow-2xl ${
              isBusy
                ? "cursor-not-allowed opacity-80"
                : "cursor-pointer hover:scale-105 transition-transform"
            }`}
            onClick={startRecording}
            disabled={isBusy}
            aria-label="رڪارڊنگ شروع ڪريو"
          >
            {audioSource === "system" ? (
              <svg className="w-16 h-16 fill-[#0474ff]" viewBox="0 0 24 24">
                <path d="M20 18c1.1 0 1.99-.9 1.99-2L22 6c0-1.1-.9-2-2-2H4c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2H0v2h24v-2h-4zM4 6h16v10H4V6z"/>
              </svg>
            ) : (
              <svg className="w-16 h-16 fill-[#0474ff]" viewBox="0 0 24 24">
                <path d="M12 14c1.66 0 3-1.34 3-3V5c0-1.66-1.34-3-3-3S9 3.34 9 5v6c0 1.66 1.34 3 3 3z" />
                <path d="M17 11c0 2.76-2.24 5-5 5s-5-2.24-5-5H5c0 3.53 2.61 6.43 6 6.92V21h2v-3.08c3.39-.49 6-3.39 6-6.92h-2z" />
              </svg>
            )}
            <span className="text-[10px] font-bold text-[#0474ff] mt-1">
              {audioSource === "system" ? "سسٽم آڊيو" : "مائيڪرو فون"}
            </span>
          </button>
        </div>

        <h3
          className={`text-lg font-semibold transition-all duration-300 text-center ${
            isBusy ? "opacity-100 translate-y-0" : "opacity-0 -translate-y-4"
          }`}
        >
          {isAnalyzing
            ? "گيت جي ڳولا جاري آهي..."
            : audioSource === "system"
            ? "ڪمپيوٽر جو آواز پڪڙي رهيو آهي..."
            : "موسيقي ٻڌي رهيو آهي..."}
        </h3>
      </div>

      {/* Catalog Modal */}
      {showCatalog && (
        <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white text-gray-900 rounded-2xl w-full max-w-lg p-6 shadow-2xl max-h-[85vh] flex flex-col">
            <div className="flex justify-between items-center border-b pb-3 mb-4">
              <div>
                <h4 className="text-xl font-bold text-gray-900">ڊيٽابيس جا سنڌي گيت</h4>
                <p className="text-xs text-gray-500">MySQL / phpMyAdmin (800,000+ Fingerprints)</p>
              </div>
              <button
                onClick={() => setShowCatalog(false)}
                className="text-gray-400 hover:text-gray-700 font-bold text-xl px-2"
              >
                ✕
              </button>
            </div>

            <div className="overflow-y-auto flex-1 space-y-3 pr-1">
              {isLoadingSongs ? (
                <div className="text-center py-8 text-gray-500">لوڊ ٿي رهيو آهي...</div>
              ) : songsList.length === 0 ? (
                <div className="text-center py-8 text-gray-500">ڪو به گيت موجود ناهي</div>
              ) : (
                songsList.map((song, idx) => (
                  <div
                    key={song.id || idx}
                    className="p-3.5 bg-slate-50 hover:bg-slate-100 rounded-xl border border-slate-200 transition-all flex items-center gap-3"
                  >
                    <div className="w-10 h-10 rounded-full bg-[#0474ff]/10 text-[#0474ff] font-bold flex items-center justify-center shrink-0">
                      {idx + 1}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="font-bold text-gray-800 text-base truncate">{song.title}</div>
                      <div className="text-xs text-gray-500 truncate">{song.artist}</div>
                    </div>
                    <span className="text-xs bg-emerald-100 text-emerald-800 px-2 py-1 rounded-full font-semibold">
                      محفوظ
                    </span>
                  </div>
                ))
              )}
            </div>

            <div className="mt-4 pt-3 border-t flex justify-end">
              <button
                onClick={() => setShowCatalog(false)}
                className="bg-gray-100 hover:bg-gray-200 text-gray-800 font-semibold px-4 py-2 rounded-lg text-sm transition-all"
              >
                بند ڪريو
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default AudioRecorder;
