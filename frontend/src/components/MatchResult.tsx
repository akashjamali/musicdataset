import { useEffect, useState } from "react";

export type SongRecord = {
  id?: number;
  title?: string;
  artist?: string;
  artists?: string[] | string | { String?: string; Valid?: boolean };
  album?: string | { String?: string; Valid?: boolean };
  image_url?: string | { String?: string; Valid?: boolean };
  duration_ms?: number | { Int32?: number; Valid?: boolean };
  created_at?: string;
};

export type Match = SongRecord;

interface MatchResultProps {
  match: SongRecord;
  onReset: () => void;
}

const safeStr = (val: unknown): string => {
  if (!val) return "";
  if (typeof val === "string") return val;
  if (typeof val === "object" && val !== null && "String" in val) {
    const obj = val as { String?: string; Valid?: boolean };
    return obj.Valid !== false ? obj.String || "" : "";
  }
  return String(val);
};

const safeArtists = (val: unknown): string => {
  if (!val) return "";
  if (Array.isArray(val)) {
    return val.map((v) => safeStr(v)).filter(Boolean).join("، ");
  }
  if (typeof val === "string") {
    try {
      const parsed = JSON.parse(val);
      if (Array.isArray(parsed)) return parsed.map((v) => safeStr(v)).join("، ");
    } catch {
      return val;
    }
    return val;
  }
  return safeStr(val);
};

const MatchResult = ({ match, onReset }: MatchResultProps) => {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    setIsVisible(true);
  }, []);

  const title = safeStr(match?.title);
  const artists = safeArtists(match?.artists);
  const album = safeStr(match?.album);
  const imageUrl = safeStr(match?.image_url);

  return (
    <div
      dir="rtl"
      className={`overflow-hidden absolute inset-0 z-50 flex flex-col items-center justify-between bg-black/90 text-white p-6 transition-opacity duration-1000 ${
        isVisible ? "opacity-100" : "opacity-0"
      }`}
    >
      {imageUrl ? (
        <div
          className="absolute inset-0 bg-cover bg-center opacity-40 blur-3xl scale-125 z-0"
          style={{ backgroundImage: `url(${imageUrl})` }}
        />
      ) : (
        <div className="absolute inset-0 bg-gradient-to-b from-blue-900/40 via-black to-black z-0" />
      )}
      <div className="absolute inset-0 bg-gradient-to-b from-black/30 via-transparent to-black/90 z-0" />

      <div className="relative z-10 w-full flex justify-between items-center pt-4">
        <button
          onClick={onReset}
          aria-label="واپس وڃو"
          className="p-2 rounded-full bg-white/10 backdrop-blur-md hover:bg-white/20 transition-all cursor-pointer"
        >
          <svg
            className="w-6 h-6 text-white"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 5l7 7-7 7"
            />
          </svg>
        </button>
        <span className="text-sm font-medium text-white/70">گيت سڃاڻپ نتيجو</span>
        <div className="w-10" />
      </div>

      <div className="relative z-10 flex flex-col items-center w-full max-w-sm space-y-6 my-auto">
        <div
          className={`relative w-64 h-64 rounded-2xl overflow-hidden shadow-2xl shadow-blue-500/20 transition-all duration-1000 delay-300 bg-white/5 flex items-center justify-center ${
            isVisible ? "translate-y-0 opacity-100" : "translate-y-10 opacity-0"
          }`}
        >
          {imageUrl ? (
            <img
              src={imageUrl}
              alt={album || title}
              className="w-full h-full object-cover"
            />
          ) : (
            <div className="flex flex-col items-center justify-center text-blue-400">
              <svg className="w-24 h-24" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 3v10.55c-.59-.34-1.27-.55-2-.55-2.21 0-4 1.79-4 4s1.79 4 4 4 4-1.79 4-4V7h4V3h-6z" />
              </svg>
            </div>
          )}
        </div>

        <div
          className={`text-center space-y-2 transition-all duration-1000 delay-500 ${
            isVisible ? "translate-y-0 opacity-100" : "translate-y-10 opacity-0"
          }`}
        >
          <h1 className="text-3xl font-bold leading-tight line-clamp-2">
            {title || "اڻڄاتل گيت"}
          </h1>
          {artists && (
            <p className="text-xl text-blue-400 font-medium">
              {artists}
            </p>
          )}
          {album && (
            <p className="text-sm text-gray-400 uppercase tracking-wider mt-1">
              {album}
            </p>
          )}
        </div>
      </div>

      <div
        className={`relative z-10 w-full max-w-sm space-y-3 pb-8 transition-all duration-1000 delay-700 ${
          isVisible ? "translate-y-0 opacity-100" : "translate-y-10 opacity-0"
        }`}
      >
        <button
          onClick={onReset}
          className="w-full py-3.5 bg-white/10 hover:bg-white/20 backdrop-blur-md text-white font-semibold rounded-full transition-transform active:scale-95 cursor-pointer border border-white/15 flex items-center justify-center gap-2"
        >
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          ٻيو گيت سڃاڻو
        </button>
      </div>
    </div>
  );
};

export default MatchResult;
