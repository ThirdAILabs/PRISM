#pragma once

#include <algorithm>
#include <memory>
#include <optional>
#include <string>
#include <unordered_map>
#include <vector>

namespace thirdai::automl::udt {

class Flash {
public:
  static std::unique_ptr<Flash>
  make(const std::optional<std::string> &input_column,
       const std::string &target_column, const std::string &dataset_size);

  void trainOnFile(const std::string &filename);

  std::vector<std::string> predictSimple(const std::string &sample,
                                         uint32_t top_k);
};

} // namespace thirdai::automl::udt